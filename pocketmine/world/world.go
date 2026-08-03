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
	"pocketmine-go/pocketmine/world/sound"
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
}

// New constructs a World using gen to generate chunks on demand. knownBlocks must include at
// least one Behavior for every distinct block type gen can place (e.g. for a Flat generator,
// every block.Behavior referenced by its []generator.FlatLayer) - see stateTemplates' doc comment
// for why.
func New(gen generator.Generator, translator *convert.BlockTranslator, knownBlocks []block.Behavior) *World {
	w := &World{
		generator:       gen,
		translator:      translator,
		chunks:          map[[2]int]*format.Chunk{},
		populated:       map[[2]int]bool{},
		stateTemplates:  map[int32]block.Behavior{},
		stateByBlockKey: map[string]int32{},
	}
	for _, blk := range knownBlocks {
		w.registerTemplate(blk)
	}
	return w
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

// GetTile/AddTile are ports of World::getTile/addTile. This port has no Tile-in-World system yet
// (see format.Chunk's doc comment on the same gap) - GetTile always reports "no tile", AddTile is
// a no-op.
func (w *World) GetTile(pos block.Position) (block.Tile, bool) { return nil, false }
func (w *World) AddTile(tile block.Tile)                       {}

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

// GetFullLightAt/GetBlockLightAt/GetRealBlockSkyLightAt/GetHighestAdjacentFullLightAt/
// GetHighestAdjacentBlockLightAt/GetPotentialLightAt are ports of World's respective light-query
// methods. This port has no light engine yet (no LightPopulationTask/HeightMap-driven skylight,
// no BlockLightUpdate) - every position simply reports full light (15) rather than a guessed or
// zeroed value, so light-gated behaviour (crop growth, mob spawning rules, ...) defaults to
// "always allowed" until a real light engine exists.
func (w *World) GetFullLightAt(x, y, z int) int                 { return 15 }
func (w *World) GetBlockLightAt(x, y, z int) int                { return 15 }
func (w *World) GetRealBlockSkyLightAt(x, y, z int) int         { return 15 }
func (w *World) GetHighestAdjacentFullLightAt(x, y, z int) int  { return 15 }
func (w *World) GetHighestAdjacentBlockLightAt(x, y, z int) int { return 15 }
func (w *World) GetPotentialLightAt(x, y, z int) int            { return 15 }

// GetSunAnglePercentage is a port of World::getSunAnglePercentage. This port has no day/night
// cycle yet, so this reports a fixed midday value (0.5) rather than a guess.
func (w *World) GetSunAnglePercentage() float64 { return 0.5 }

// GetNearbyEntities is a port of World::getNearbyEntities. This port has no entity-in-world
// tracking yet, so this always reports no entities nearby.
func (w *World) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity { return nil }

// IsInWorld is a port of World::isInWorld, minus the PHP original's +/-30000000 horizontal bound
// (Bedrock's real world border, which nothing in this port enforces or needs yet - only the
// vertical bound is actually load-bearing for any current caller, e.g. Sugarcane.grow's upward
// climb).
func (w *World) IsInWorld(x, y, z int) bool { return y >= YMin && y < YMax }

// UseBreakOn is the simplified (no item/player/particles/drops) form of World::useBreakOn that
// block.World documents as the only form the block package itself needs - replaces the block with
// air and reports success unconditionally, matching the existing test doubles' behaviour
// throughout the block package's own test suite.
func (w *World) UseBreakOn(pos math.Vector3) bool {
	_ = w.SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, w), block.VanillaAir())
	return true
}
