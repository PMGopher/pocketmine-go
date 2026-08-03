// Package world is a port of a slice of pocketmine\world\World, including a real on-disk
// LevelDB-backed WorldProvider (see OpenProvider/SaveAll) - not in-memory-only any more.
package world

import (
	"fmt"
	"sort"
	"strings"

	goleveldb "github.com/syndtr/goleveldb/leveldb"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/format"
	worldio "pocketmine-go/pocketmine/world/format/io/leveldb"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/light"
	"pocketmine-go/pocketmine/world/sound"
	"pocketmine-go/pocketmine/world/utils"
)

var _ block.World = (*World)(nil)

// YMin/YMax mirror pocketmine\world\World::Y_MIN/Y_MAX for the subchunk range format.Chunk
// actually stores (see format.Chunk's own MinSubChunkIndex/MaxSubChunkIndex).
const (
	YMin = format.MinSubChunkIndex * format.SubChunkEdgeLength
	YMax = (format.MaxSubChunkIndex + 1) * format.SubChunkEdgeLength
)

// World is a minimal port of pocketmine\world\World: enough to generate chunks on demand and let
// block.Behavior read/write blocks through the block.World interface it already codes against.
// Persistence (loading/saving to disk - this port has no WorldProvider), real light calculation
// (LightPopulationTask/BlockLightUpdate), entity tracking, and tile support are all documented
// gaps on the methods below, rather than silently wrong behaviour - every one returns an honest,
// clearly-labelled placeholder instead of guessing.
type World struct {
	generator  generator.Generator
	translator *convert.BlockTranslator

	chunks map[[2]int]*format.Chunk

	// populated tracks which chunks have already run PopulateChunk - see ensurePopulated's doc
	// comment for why this is needed on top of chunks' own presence check.
	populated map[[2]int]bool

	// stateTemplates lets GetBlockAt reconstruct a block.Behavior from the bare internal state ID
	// a Chunk stores (Chunk deliberately never keeps live Behavior instances around - see
	// format.Chunk's doc comment on why). Every block type the World needs to be able to read back
	// must be registered up front (see New's knownBlocks parameter); SetBlock also self-registers
	// whatever it's given, so anything ever placed through it becomes readable too.
	stateTemplates map[int32]block.Behavior

	// stateByBlockKey is stateTemplates' reverse direction (persistent name+states -> internal
	// state ID) - built alongside it in registerTemplate, needed to resolve a block state read
	// back from a saved world file (see leveldb.StateResolver). Keyed by blockStateKey.
	stateByBlockKey map[string]int32

	// provider is this World's on-disk backing (see OpenProvider) - nil means pure in-memory, no
	// different from this port's original design (every chunk generated fresh, nothing survives a
	// restart).
	provider *goleveldb.DB

	// lightFilters/lightEmitters/directSkyLightBlockers are the light engine's per-block-state
	// lookup tables (see pocketmine/world/light), built in registerTemplate from every registered
	// block's real GetLightFilter()/GetLightLevel()/BlocksDirectSkyLight() - the same
	// "only what's registered" scope every other per-state table in this port already has.
	// lightFilters is min(15, GetLightFilter()+light.BaseLightFilter), matching
	// RuntimeBlockStateRegistry.php's own formula (see light package's doc comment on why the
	// raw GetLightFilter() alone would be wrong - it would make light never attenuate through
	// air at all).
	lightFilters           map[int32]int
	lightEmitters          map[int32]int
	directSkyLightBlockers map[int32]bool

	subChunkExplorer  *utils.SubChunkExplorer
	skyLightUpdate    *light.SkyLightUpdate
	blockLightUpdate  *light.BlockLightUpdate
	skyLightReduction int // see GetSkyLightReduction's doc comment

	// entities is this World's entity registry (see AddEntity/RemoveEntity/GetNearbyEntities) -
	// keyed by registeredEntity.GetID(), a flat map rather than PHP's per-chunk index (see
	// GetNearbyEntities' own doc comment on why that's a performance detail, not a correctness one).
	entities map[int]registeredEntity
}

// New constructs a World using gen to generate chunks on demand. knownBlocks must include at
// least one Behavior for every distinct block type gen can place (e.g. for a Flat generator,
// every block.Behavior referenced by its []generator.FlatLayer) - see stateTemplates' doc comment
// for why.
func New(gen generator.Generator, translator *convert.BlockTranslator, knownBlocks []block.Behavior) *World {
	w := &World{
		generator:              gen,
		translator:             translator,
		chunks:                 map[[2]int]*format.Chunk{},
		populated:              map[[2]int]bool{},
		stateTemplates:         map[int32]block.Behavior{},
		stateByBlockKey:        map[string]int32{},
		lightFilters:           map[int32]int{},
		lightEmitters:          map[int32]int{},
		directSkyLightBlockers: map[int32]bool{},
		entities:               map[int]registeredEntity{},
	}
	w.subChunkExplorer = utils.NewSubChunkExplorer(w)
	w.skyLightUpdate = light.NewSkyLightUpdate(w.subChunkExplorer, w.lightFilters, w.directSkyLightBlockers)
	w.blockLightUpdate = light.NewBlockLightUpdate(w.subChunkExplorer, w.lightFilters, w.lightEmitters)

	for _, blk := range knownBlocks {
		w.registerTemplate(blk)
	}
	return w
}

// GetChunk is a port of World::getChunk: a non-generating lookup (unlike GetOrLoadChunk/
// generateChunkOnly, this never creates a chunk that isn't already loaded) - the
// utils.ChunkSource capability SubChunkExplorer (and so the light engine) needs.
func (w *World) GetChunk(chunkX, chunkZ int) (*format.Chunk, bool) {
	c, ok := w.chunks[chunkKey(chunkX, chunkZ)]
	return c, ok
}

// registerTemplate records blk as the reconstruction template for its state ID and warms the
// BlockTranslator's cache for the same ID (see BlockTranslator.NetworkIDForCachedState's doc
// comment on why chunk serialization needs that cache pre-warmed). Also records the reverse
// (persistent name+states -> state ID) mapping a saved world's LoadChunk needs (see
// stateByBlockKey's doc comment) - blocks with no registered BlockStateSerializer yet simply
// can't be saved/loaded correctly, the same "not supported over the network yet" gap
// InternalIDToNetworkID already has, just for disk instead of network.
func (w *World) registerTemplate(blk block.Behavior) {
	stateID := int32(blk.GetStateId())
	if _, ok := w.stateTemplates[stateID]; !ok {
		w.stateTemplates[stateID] = blk.Clone()
	}
	w.translator.InternalIDToNetworkID(blk)

	if data, err := convert.SerializeBlockState(blk); err == nil {
		if _, exists := w.stateByBlockKey[blockStateKey(data)]; !exists {
			w.stateByBlockKey[blockStateKey(data)] = stateID
		}
	}

	if _, ok := w.lightFilters[stateID]; !ok {
		w.lightFilters[stateID] = min(15, blk.GetLightFilter()+light.BaseLightFilter)
		w.lightEmitters[stateID] = blk.GetLightLevel()
		w.directSkyLightBlockers[stateID] = blk.BlocksDirectSkyLight()
	}
}

// blockStateKey builds a deterministic string key from a bedrock.BlockStateData's name and
// states, for use as a map key (bedrock.BlockStateData itself isn't comparable - States is a
// map). State property values are always int32, uint8 or string (see
// convert/vanilla_block_mappings.go's registrations) - anything else is a programmer error there,
// not something this needs to handle gracefully.
func blockStateKey(data bedrock.BlockStateData) string {
	names := make([]string, 0, len(data.States))
	for name := range data.States {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(data.Name)
	for _, name := range names {
		fmt.Fprintf(&b, ";%s=%T:%v", name, data.States[name], data.States[name])
	}
	return b.String()
}

func chunkKey(chunkX, chunkZ int) [2]int { return [2]int{chunkX, chunkZ} }

// GetOrLoadChunk returns the chunk at the given chunk coordinates, generating and populating (see
// Generator.PopulateChunk) it via the configured Generator on first access - this port has no
// on-disk world storage (no WorldProvider equivalent), so "load" always means "generate". This is
// the entry point external callers (main.go, tests) should use to get a chunk ready to look at or
// send to a client; code that runs *during* population itself (GetBlockAt/SetBlock/
// GetOrLoadChunkAtPosition, used by populators reading/writing across chunk borders) deliberately
// goes through generateChunkOnly instead - see ensurePopulated's doc comment for why.
func (w *World) GetOrLoadChunk(chunkX, chunkZ int) *format.Chunk {
	c := w.generateChunkOnly(chunkX, chunkZ)
	w.ensurePopulated(chunkX, chunkZ)
	return c
}

// generateChunkOnly returns the chunk at the given coordinates: from cache if already loaded,
// else from the on-disk provider if one is open and has this chunk saved (see OpenProvider),
// else generating (and caching) it via the configured Generator - but never running population,
// see ensurePopulated's doc comment for why callers reached from inside a populate pass must use
// this instead of GetOrLoadChunk.
func (w *World) generateChunkOnly(chunkX, chunkZ int) *format.Chunk {
	key := chunkKey(chunkX, chunkZ)
	if c, ok := w.chunks[key]; ok {
		return c
	}

	if w.provider != nil {
		if c, ok, err := worldio.LoadChunk(w.provider, int32(chunkX), int32(chunkZ), int32(block.VanillaAir().GetStateId()), 0, w.resolveBlockState); err == nil && ok {
			w.chunks[key] = c
			w.populated[key] = true // a saved chunk was always fully generated+populated before being saved
			return c
		}
	}

	c := w.generator.GenerateChunk(chunkX, chunkZ)
	w.chunks[key] = c
	return c
}

// resolveBlockState adapts stateByBlockKey to leveldb.StateResolver's shape.
func (w *World) resolveBlockState(data bedrock.BlockStateData) (int32, bool) {
	stateID, ok := w.stateByBlockKey[blockStateKey(data)]
	return stateID, ok
}

// lookupBlockState adapts stateTemplates to leveldb.StateLookup's shape.
func (w *World) lookupBlockState(stateID int32) (bedrock.BlockStateData, error) {
	tpl, ok := w.stateTemplates[stateID]
	if !ok {
		return bedrock.BlockStateData{}, fmt.Errorf("world: no registered template for internal state %d", stateID)
	}
	return convert.SerializeBlockState(tpl)
}

// OpenProvider opens (creating if necessary) a real Bedrock-compatible LevelDB world database at
// path, and attaches it to this World as its on-disk backing - a port of the relevant slice of
// WorldManager::loadWorld/World::__construct's WorldProvider setup. Chunks are loaded from it
// lazily (see generateChunkOnly) and must be explicitly written back with SaveAll (this port has
// no per-chunk dirty tracking / autosave scheduler yet, so saving is all-or-nothing and
// caller-triggered - see main.go's shutdown handler).
func (w *World) OpenProvider(path string) error {
	db, err := goleveldb.OpenFile(path, nil)
	if err != nil {
		return fmt.Errorf("world: opening LevelDB world at %q: %w", path, err)
	}
	w.provider = db
	return nil
}

// SaveAll writes every currently-loaded chunk back to the open provider (see OpenProvider) - a
// no-op if no provider is open.
func (w *World) SaveAll() error {
	if w.provider == nil {
		return nil
	}
	for key, chunk := range w.chunks {
		if err := worldio.SaveChunk(w.provider, int32(key[0]), int32(key[1]), chunk, w.lookupBlockState); err != nil {
			return err
		}
	}
	return nil
}

// Close saves every loaded chunk (see SaveAll) and closes the on-disk provider, if one is open.
func (w *World) Close() error {
	if w.provider == nil {
		return nil
	}
	if err := w.SaveAll(); err != nil {
		return err
	}
	return w.provider.Close()
}

// ensurePopulated runs this chunk's Populators exactly once (see the populated map), first making
// sure its 8 immediate neighbours are generated - not populated - via generateChunkOnly, so a
// populator can safely read/write a handful of blocks across a chunk border (matching Ore's blast
// radius) without that write recursively triggering the neighbour's own population.
//
// Real PocketMine-MP achieves the same guarantee differently: it defers a chunk's PopulationTask
// until World::orderChunkPopulation sees all 8 neighbours already generated, running generation and
// population as separate asynchronous passes over a whole neighbourhood. This port has no
// worker-thread/task-queue system to defer onto, so it does both synchronously and immediately
// on first access instead - but still keeps "generate a neighbour" and "populate a neighbour"
// as two distinct steps, which is the part that actually matters: collapsing them into one (as an
// earlier version of this method did) means populating chunk (0,0) can write into chunk (1,0),
// whose own populate call can reach into (2,0), and so on - an unbounded chain reaction that
// eagerly populates the entire world. Only ever generating (never populating) neighbours here is
// what stops that chain.
func (w *World) ensurePopulated(chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	if w.populated[key] {
		return
	}
	w.populated[key] = true

	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			w.generateChunkOnly(chunkX+dx, chunkZ+dz)
		}
	}
	w.generator.PopulateChunk(w, chunkX, chunkZ)

	// Light is recalculated after population (not generation) so it reflects the final terrain -
	// populators (Tree, Ore, ...) can add/remove blocks with real light filters/emission that
	// pure generation didn't have yet. Matches real PocketMine-MP's own dependency: a
	// PopulationTask always runs before a chunk is considered ready for LightPopulationTask.
	// Neighbours only being *generated* (not populated) here is exactly as safe for light
	// propagation as it already is for population's own cross-border writes - see this method's
	// own doc comment above on why that's the invariant that matters, not "fully populated".
	w.skyLightUpdate.RecalculateChunk(chunkX, chunkZ)
	w.skyLightUpdate.Execute()
	w.blockLightUpdate.RecalculateChunk(chunkX, chunkZ)
	w.blockLightUpdate.Execute()
	w.chunks[key].SetLightPopulated(true, true)
}

// Translator returns the BlockTranslator this World's chunks were populated through - the network
// chunk serializer needs it to translate a chunk's stored internal state IDs to Bedrock runtime
// IDs.
func (w *World) Translator() *convert.BlockTranslator { return w.translator }

// GetBlockAt is a port of World::getBlockAt (simplified: no chunk-load-failure/out-of-bounds
// handling - see IsInWorld). Falls back to air for a state ID with no registered template (see
// stateTemplates' doc comment) rather than panicking; this should never happen for a state this
// World itself ever wrote, only for data corruption.
func (w *World) GetBlockAt(x, y, z int) block.Behavior {
	chunk := w.generateChunkOnly(x>>4, z>>4)
	stateID := chunk.GetBlockStateID(x&0xf, y, z&0xf)
	tpl, ok := w.stateTemplates[stateID]
	if !ok {
		tpl = w.stateTemplates[int32(block.VanillaAir().GetStateId())]
	}
	got := tpl.Clone()
	got.(positionable).SetPosition(w, x, y, z)
	return got
}

// positionable is satisfied by every concrete block type via SetPosition's promotion from
// *block.Block - not part of block.Behavior itself (see this port's established convention for
// promoted-but-not-interface methods, e.g. block.asItemOrNil's identical reasoning for AsItem).
type positionable interface {
	SetPosition(world block.World, x, y, z int)
}

// SetBlock is a port of World::setBlock. Also registers blk as a state template (see
// registerTemplate) so it can be read back later even if it wasn't in New's knownBlocks list.
func (w *World) SetBlock(pos block.Position, blk block.Behavior) error {
	x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()
	w.registerTemplate(blk)
	chunk := w.generateChunkOnly(x>>4, z>>4)
	chunk.SetBlockStateID(x&0xf, y, z&0xf, int32(blk.GetStateId()))
	return nil
}

// GetTile is a port of World::getTile.
func (w *World) GetTile(pos block.Position) (block.Tile, bool) {
	return w.GetTileAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())
}

// GetTileAt is a port of World::getTileAt - also satisfies tile.World's own GetTileAt (needed by
// e.g. Chest's neighbour-pairing lookup).
func (w *World) GetTileAt(x, y, z int) (block.Tile, bool) {
	chunk := w.generateChunkOnly(x>>4, z>>4)
	return chunk.GetTile(x&0xf, y, z&0xf)
}

// AddTile is a port of World::addTile.
func (w *World) AddTile(t block.Tile) {
	pos := t.GetPosition()
	chunk := w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4)
	chunk.AddTile(t)
}

// RemoveTile is a port of World::removeTile - also satisfies tile.World's own RemoveTile (called
// by TileBase.Close() implementations, matching Tile::close()'s real `$this->lastUsedChunk?->
// removeTile($this)`-style bookkeeping).
func (w *World) RemoveTile(t block.Tile) {
	pos := t.GetPosition()
	chunk := w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4)
	chunk.RemoveTile(t)
}

// chunkAdapter satisfies block.Chunk (just SetBlockStateID(x,y,z,stateID int)) over a
// *format.Chunk (whose own SetBlockStateID takes int32 state IDs, matching PalettedBlockArray's
// storage type) - a thin type adapter, not a second implementation of anything.
type chunkAdapter struct{ chunk *format.Chunk }

func (a chunkAdapter) SetBlockStateID(x, y, z int, stateID int) {
	a.chunk.SetBlockStateID(x, y, z, int32(stateID))
}

func (a chunkAdapter) GetHighestBlockAt(x, z int) (int, bool) {
	return a.chunk.GetHighestBlockAt(x, z)
}

func (a chunkAdapter) GetBiomeID(x, y, z int) int32 {
	return a.chunk.GetBiomeID(x, y, z)
}

// GetOrLoadChunkAtPosition is a port of World::getOrLoadChunkAtPosition. Uses generateChunkOnly,
// not GetOrLoadChunk - see ensurePopulated's doc comment on why code reachable from inside a
// populate pass (this is how populator.TallGrass looks up a chunk's heightmap) must not trigger
// population itself.
func (w *World) GetOrLoadChunkAtPosition(pos block.Position) (block.Chunk, bool) {
	return chunkAdapter{w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4)}, true
}

// AddSound is a port of World::addSound. This port has no player-session/packet-broadcast system
// yet to actually deliver a sound packet to anyone, so this is a documented no-op rather than a
// guess at how broadcasting should work.
func (w *World) AddSound(pos math.Vector3, s sound.Sound) {}

// ScheduleDelayedBlockUpdate is a port of World::scheduleDelayedBlockUpdate. This port has no
// world tick scheduler yet (see pocketmine/scheduler, which exists but isn't wired to a running
// World), so this is a documented no-op.
func (w *World) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {}

// GetBlockLightAt is a port of World::getBlockLightAt.
func (w *World) GetBlockLightAt(x, y, z int) int {
	if !w.IsInWorld(x, y, z) {
		return 0
	}
	chunk, ok := w.GetChunk(x>>4, z>>4)
	if !ok {
		return 0
	}
	if populated, ok := chunk.IsLightPopulated(); !ok || !populated {
		return 0 //TODO: this should probably throw instead (light not calculated yet) - see World::getBlockLightAt's own TODO
	}
	return chunk.GetBlockLightAt(x&0xf, y, z&0xf)
}

// getPotentialBlockSkyLightAt is a port of World::getPotentialBlockSkyLightAt.
func (w *World) getPotentialBlockSkyLightAt(x, y, z int) int {
	if !w.IsInWorld(x, y, z) {
		if y >= YMax {
			return 15
		}
		return 0
	}
	chunk, ok := w.GetChunk(x>>4, z>>4)
	if !ok {
		return 0
	}
	if populated, ok := chunk.IsLightPopulated(); !ok || !populated {
		return 0
	}
	return chunk.GetBlockSkyLightAt(x&0xf, y, z&0xf)
}

// GetRealBlockSkyLightAt is a port of World::getRealBlockSkyLightAt.
func (w *World) GetRealBlockSkyLightAt(x, y, z int) int {
	return max(0, w.getPotentialBlockSkyLightAt(x, y, z)-w.skyLightReduction)
}

// GetPotentialLightAt is a port of World::getPotentialLightAt.
func (w *World) GetPotentialLightAt(x, y, z int) int {
	return max(w.getPotentialBlockSkyLightAt(x, y, z), w.GetBlockLightAt(x, y, z))
}

// GetFullLightAt is a port of World::getFullLightAt.
func (w *World) GetFullLightAt(x, y, z int) int {
	skyLight := w.GetRealBlockSkyLightAt(x, y, z)
	if skyLight < 15 {
		return max(skyLight, w.GetBlockLightAt(x, y, z))
	}
	return skyLight
}

// getHighestAdjacentLight is a port of World::getHighestAdjacentLight.
func (w *World) getHighestAdjacentLight(x, y, z int, lightGetter func(x, y, z int) int) int {
	result := 0
	for _, side := range math.AllFacing {
		off := math.FacingOffset[side]
		x1, y1, z1 := x+off[0], y+off[1], z+off[2]

		if !w.IsInWorld(x1, y1, z1) {
			continue
		}
		chunk, ok := w.GetChunk(x1>>4, z1>>4)
		if !ok {
			continue
		}
		if populated, ok := chunk.IsLightPopulated(); !ok || !populated {
			continue
		}
		if v := lightGetter(x1, y1, z1); v > result {
			result = v
		}
	}
	return result
}

// GetHighestAdjacentFullLightAt is a port of World::getHighestAdjacentFullLightAt.
func (w *World) GetHighestAdjacentFullLightAt(x, y, z int) int {
	return w.getHighestAdjacentLight(x, y, z, w.GetFullLightAt)
}

// GetHighestAdjacentBlockLightAt is a port of World::getHighestAdjacentBlockLight.
func (w *World) GetHighestAdjacentBlockLightAt(x, y, z int) int {
	return w.getHighestAdjacentLight(x, y, z, w.GetBlockLightAt)
}

// GetSkyLightReduction is a port of World::getSkyLightReduction: how much sky light is currently
// reduced by weather/time of day. This port has no real weather/day-night cycle driving it yet
// (see GetSunAnglePercentage's own doc comment on the same gap), so it's a fixed 0 (a permanently
// clear midday sky) rather than a guess - revisited once real time/weather exist.
func (w *World) GetSkyLightReduction() int { return w.skyLightReduction }

// GetSunAnglePercentage is a port of World::getSunAnglePercentage. This port has no day/night
// cycle yet, so this reports a fixed midday value (0.5) rather than a guess.
func (w *World) GetSunAnglePercentage() float64 { return 0.5 }

// registeredEntity is the minimal surface World's entity registry needs beyond block.Entity
// itself (a unique ID, and whether it's already been closed) - declared locally, matching this
// port's established forward-compatible-local-interface convention, so a future concrete Entity
// type (pocketmine/entity only has the Entity/Living markers so far - no concrete spawnable type
// exists yet to actually register) satisfies it structurally with no import needed here.
type registeredEntity interface {
	block.Entity
	GetID() int
	IsClosed() bool
}

// intersectEpsilon matches AxisAlignedBB::intersectsWith's real PHP default parameter
// ($epsilon = 0.00001).
const intersectEpsilon = 0.00001

// AddEntity is a port of World::addEntity. Panics for a closed entity or one already registered
// under a different instance for the same ID, matching the PHP original's InvalidArgumentException
// (both are programmer errors at the call site).
func (w *World) AddEntity(e registeredEntity) {
	if e.IsClosed() {
		panic("world: attempted to add a closed entity to the world")
	}
	if existing, ok := w.entities[e.GetID()]; ok && existing != e {
		panic("world: attempted to create another entity with the same ID")
	}
	w.entities[e.GetID()] = e
}

// RemoveEntity is a port of World::removeEntity.
func (w *World) RemoveEntity(e registeredEntity) {
	delete(w.entities, e.GetID())
}

// GetEntity is a port of World::getEntity.
func (w *World) GetEntity(id int) (registeredEntity, bool) {
	e, ok := w.entities[id]
	return e, ok
}

// GetEntities is a port of World::getEntities.
func (w *World) GetEntities() map[int]registeredEntity { return w.entities }

// GetNearbyEntities is a port of World::getNearbyEntities (minus its $entity exclusion parameter,
// which block.World's interface doesn't need - nothing in the block package calls this with an
// entity to exclude). Unlike the PHP original, this doesn't pre-filter by nearby chunks first (a
// pure performance optimisation over exactly the same correct result set, not a correctness
// requirement) - this port has no per-chunk entity index (see registeredEntity's own doc comment
// on why there's nothing to spawn into one yet anyway), so a flat scan of every registered entity
// produces an identical result, just O(entity count) instead of O(nearby chunks' entity count).
func (w *World) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity {
	var nearby []block.Entity
	for _, e := range w.entities {
		if e.GetBoundingBox().IntersectsWith(bb, intersectEpsilon) {
			nearby = append(nearby, e)
		}
	}
	return nearby
}

// int32Min/int32Max mirror pocketmine\utils\Limits::INT32_MIN/INT32_MAX - named locally instead
// of using Go's stdlib math.MinInt32/MaxInt32 since this file already imports this port's own
// math package under the plain "math" name.
const (
	int32Min = -1 << 31
	int32Max = 1<<31 - 1
)

// IsInWorld is a port of World::isInWorld: x/z must be within the int32 range (Limits::INT32_MIN/
// MAX - not a "world border" game feature, just the same defensive coordinate-overflow bound real
// PocketMine-MP enforces), and y within this World's vertical bounds.
func (w *World) IsInWorld(x, y, z int) bool {
	return x >= int32Min && x <= int32Max &&
		y >= YMin && y < YMax &&
		z >= int32Min && z <= int32Max
}

// UseBreakOn is the simplified (no item/player/particles/drops) form of World::useBreakOn that
// block.World documents as the only form the block package itself needs - replaces the block with
// air and reports success unconditionally, matching the existing test doubles' behaviour
// throughout the block package's own test suite.
func (w *World) UseBreakOn(pos math.Vector3) bool {
	_ = w.SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, w), block.VanillaAir())
	return true
}
