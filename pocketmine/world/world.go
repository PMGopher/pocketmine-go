// Package world is a port of a slice of pocketmine\world\World, including a real on-disk
// LevelDB-backed WorldProvider (see OpenProvider/SaveAll) - not in-memory-only any more.
package world

import (
	"fmt"
	"math/rand"
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	worldevent "pocketmine-go/pocketmine/event/world"
	"sort"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	goleveldb "github.com/syndtr/goleveldb/leveldb"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/promise"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/format"
	worldformatio "pocketmine-go/pocketmine/world/format/io"
	worldio "pocketmine-go/pocketmine/world/format/io/leveldb"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/light"
	"pocketmine-go/pocketmine/world/particle"
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

	// changedBlocks is World::$changedBlocks: blocks set since the last tick, per chunk, sent to
	// the players using those chunks at the end of the tick (see sendChangedBlocks).
	changedBlocks map[[2]int]map[[3]int]math.Vector3
	// sendTimeTicker is World::$sendTimeTicker: the time is sent to players every 200 ticks.
	sendTimeTicker int
	// logger is World::$logger (nil for worlds made directly with New, e.g. in tests).
	logger log.Logger

	// Asynchronous population (see world_population.go): World::$generatorExecutor,
	// $workerPool, $chunkLock, $chunkPopulationRequestMap/Queue/QueueIndex,
	// $activeChunkPopulationTasks, $maxConcurrentChunkPopulationTasks and
	// $knownUngeneratedChunks.
	generatorExecutor                 GeneratorExecutor
	workerPool                        *scheduler.AsyncPool
	chunkLock                         map[[2]int]*ChunkLockID
	chunkPopulationRequestMap         map[[2]int]*promise.Resolver[*format.Chunk]
	chunkPopulationRequestQueue       [][2]int
	chunkPopulationRequestQueueIndex  map[[2]int]bool
	activeChunkPopulationTasks        map[[2]int]bool
	maxConcurrentChunkPopulationTasks int
	knownUngeneratedChunks            map[[2]int]bool

	// registryVersion counts changes to the per-state tables; registrySnap is the frozen copy of
	// them workers use (see registrySnapshot).
	registryVersion     int
	registrySnapVersion int
	registrySnap        *blockStateRegistry

	// id/folderName/displayName are set by WorldManager (LoadWorld/GenerateWorld) - a bare in-
	// memory World constructed directly via New (as main.go's own single-world setup still does)
	// simply never has them populated, matching a world with id 0 and no name being harmless
	// (nothing in World itself reads them).
	id          int
	folderName  string
	displayName string

	// spawnLocation mirrors the world spawn position real PHP stores in its WorldProvider's
	// WorldData (see GetSpawnLocation/SetSpawnLocation) - held directly on World here instead,
	// since this port's World doesn't require an on-disk provider to exist (OpenProvider is
	// optional - see its own doc comment) the way real PHP's World::getSpawnLocation always does.
	spawnLocation math.Vector3

	// biomeRegistry backs GetBiome - a fixed ID->Biome lookup table, not generator-specific data
	// (unlike e.g. Normal's own *biome.Registry, which additionally drives pickBiome's noise-based
	// selection) - matches real PHP's BiomeRegistry::getInstance() being a single shared singleton
	// regardless of which world/generator is asking.
	biomeRegistry *biome.Registry

	chunks map[[2]int]*format.Chunk

	// stateTemplates lets GetBlockAt reconstruct a block.Behavior from the bare internal state ID
	// a Chunk stores (Chunk deliberately never keeps live Behavior instances around - see
	// format.Chunk's doc comment on why). Every block type the World needs to be able to read back
	// must be registered up front (see New's knownBlocks parameter); SetBlock also self-registers
	// whatever it's given, so anything ever placed through it becomes readable too.
	stateTemplates map[int32]block.Behavior

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

	// blastResistance is Explosion's per-state lookup table (see explosion.go), built in
	// registerTemplate from every registered block's real GetBreakInfo().GetBlastResistance() -
	// the same "only what's registered" scope every other per-state table in this port already has
	// (matches RuntimeBlockStateRegistry::$blastResistance).
	blastResistance map[int32]float64

	subChunkExplorer  *utils.SubChunkExplorer
	skyLightUpdate    *light.SkyLightUpdate
	blockLightUpdate  *light.BlockLightUpdate
	skyLightReduction int // see GetSkyLightReduction's doc comment

	// entities/entitiesByChunk/entityLastKnownPositions/updateEntities are World's entity registry
	// (see entity.go) - PHP's $entities, $entitiesByChunk, $entityLastKnownPositions and
	// $updateEntities.
	entities                 map[int]Entity
	entitiesByChunk          map[[2]int]map[int]Entity
	entityLastKnownPositions map[int]math.Vector3
	updateEntities           orderedEntitySet

	// sleepTicks is World::$sleepTicks: ticks until the sleep check (CheckSleep) runs.
	sleepTicks int
	// closed is set once the world is unloaded.
	closed bool

	// difficulty is World::getDifficulty's value. PHP reads it through the provider's WorldData
	// (level.dat); this port's World doesn't own its WorldData (cmd/pocketmine-go does), so the
	// value is held here and kept in sync by the owner - see SetDifficulty.
	difficulty int

	// --- tick.go: time/weather, scheduled+neighbour block updates, chunk loading/ticking/unload ---

	// time is a port of World::$time (whole ticks elapsed, driving day/night - see
	// computeSunAnglePercentage). Go's int64 increment wraps on overflow exactly the way real
	// PHP's actuallyDoTick explicitly simulates (PHP_INT_MAX -> PHP_INT_MIN) - not a coincidence
	// this needs replicating by hand, just two languages agreeing on twos-complement wraparound.
	time               int64
	sunAnglePercentage float64

	// randomTickBlocks is the per-state-ID "does this state tick randomly" lookup tickChunk samples
	// against - built alongside stateTemplates in registerTemplate (matches
	// RuntimeBlockStateRegistry building $this->randomTickBlocks the same way, from every
	// registered state's own ticksRandomly()).
	randomTickBlocks map[int32]bool

	// scheduledUpdates/scheduledUpdateDelay are ScheduleDelayedBlockUpdate's queue - a plain slice
	// scanned each tick (see updateScheduledBlocks) rather than PHP's SplPriorityQueue, since this
	// port has no need for the queue to stay sorted between ticks (a full scan of "is it due yet"
	// is simpler and just as correct at this port's scale). scheduledUpdateDelay mirrors
	// scheduledBlockUpdateQueueIndex - the raw requested delay, used for the "don't schedule a
	// later duplicate" dedup check.
	scheduledUpdates     []scheduledBlockUpdate
	scheduledUpdateDelay map[[3]int]int

	// neighbourUpdateQueue/neighbourUpdateQueued are NotifyNeighbourBlockUpdate's queue - same
	// plain-slice-plus-dedup-set shape as the scheduled update queue above, replacing PHP's
	// SplQueue+blockHash index with a directly-keyed [3]int (see world/light's own doc comment on
	// this port's established preference for using Go's native comparable-struct map keys instead
	// of porting World::blockHash's morton-code packing, which exists only because PHP arrays can't
	// key on a tuple).
	neighbourUpdateQueue  [][3]int
	neighbourUpdateQueued map[[3]int]bool

	// currentTick is set at the start of each DoTick call - ScheduleDelayedBlockUpdate needs it to
	// compute a due tick (currentTick+delay) even when called from outside DoTick itself (e.g. a
	// block's OnRandomTick or OnScheduledUpdate scheduling a follow-up update mid-tick).
	currentTick int64

	// chunkLoaders/tickingChunks are ChunkLoader/ChunkTicker registration - both real PHP types are
	// empty marker interfaces (zero methods - see ChunkLoader.php/ChunkTicker.php), so Go needs no
	// interface for them at all: just `any`, tracked by pointer identity (Go maps key pointers by
	// identity natively) exactly like PHP's own spl_object_id-keyed inner arrays.
	chunkLoaders  map[[2]int]map[any]bool
	tickingChunks map[[2]int]map[any]bool

	// tickRateTime is World::$tickRateTime: how long the last tick took, in milliseconds.
	tickRateTime float64

	// chunkTickRadius mirrors World::$chunkTickRadius (pocketmine.yml's chunk-ticking.tick-radius,
	// default 4) - tickChunks() is a no-op while this is <= 0.
	chunkTickRadius int

	// unloadQueue records the tick each now-loader-free chunk became eligible for unloading (see
	// UnregisterChunkLoader) - unloadChunks() only actually unloads a chunk once it's been
	// loader-free for unloadGraceTicks, matching real PHP's own 30-real-second grace window (see
	// World::unloadChunks' own `$time > ($now - 30)` check) translated to a tick count at this
	// port's fixed 20 TPS.
	unloadQueue map[[2]int]int64

	// chunkListeners backs RegisterChunkListener/UnregisterChunkListener/GetChunkListeners (see
	// chunk_listener.go) - real ChunkListener callbacks, not identity-only tracking like
	// chunkLoaders/tickingChunks above, since this one actually needs to call back into caller code.
	chunkListeners map[[2]int]map[ChunkListener]bool

	// stopTime mirrors World::$stopTime - see StopTime/StartTime's own doc comment.
	stopTime bool

	// doingTick mirrors World::$doingTick - see IsDoingTick's own doc comment.
	doingTick bool

	rng *rand.Rand
}

// unloadGraceTicks mirrors World::unloadChunks' 30-second grace window at a fixed 20 ticks/second
// (this port has no variable-TPS concept to measure real elapsed seconds against, unlike PHP's
// microtime(true) check).
const unloadGraceTicks = 30 * 20

// New constructs a World using gen to generate chunks on demand. knownBlocks must include at
// least one Behavior for every distinct block type gen can place (e.g. for a Flat generator,
// every block.Behavior referenced by its []generator.FlatLayer) - see stateTemplates' doc comment
// for why.
func New(gen generator.Generator, translator *convert.BlockTranslator, knownBlocks []block.Behavior) *World {
	w := &World{
		generator:  gen,
		translator: translator,
		chunks:     map[[2]int]*format.Chunk{},

		chunkLock:                         map[[2]int]*ChunkLockID{},
		chunkPopulationRequestMap:         map[[2]int]*promise.Resolver[*format.Chunk]{},
		chunkPopulationRequestQueueIndex:  map[[2]int]bool{},
		activeChunkPopulationTasks:        map[[2]int]bool{},
		maxConcurrentChunkPopulationTasks: defaultMaxConcurrentChunkPopulationTasks,
		knownUngeneratedChunks:            map[[2]int]bool{},
		stateTemplates:                    map[int32]block.Behavior{},
		lightFilters:                      map[int32]int{},
		lightEmitters:                     map[int32]int{},
		directSkyLightBlockers:            map[int32]bool{},
		blastResistance:                   map[int32]float64{},
		entities:                          map[int]Entity{},
		entitiesByChunk:                   map[[2]int]map[int]Entity{},
		entityLastKnownPositions:          map[int]math.Vector3{},
		difficulty:                        DifficultyNormal,
		sunAnglePercentage:                0.5,
		randomTickBlocks:                  map[int32]bool{},
		scheduledUpdateDelay:              map[[3]int]int{},
		neighbourUpdateQueued:             map[[3]int]bool{},
		chunkLoaders:                      map[[2]int]map[any]bool{},
		tickingChunks:                     map[[2]int]map[any]bool{},
		chunkTickRadius:                   4,
		unloadQueue:                       map[[2]int]int64{},
		chunkListeners:                    map[[2]int]map[ChunkListener]bool{},
		biomeRegistry:                     biome.NewRegistry(),
		rng:                               rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	w.subChunkExplorer = utils.NewSubChunkExplorer(w)
	w.skyLightUpdate = light.NewSkyLightUpdate(w.subChunkExplorer, w.lightFilters, w.directSkyLightBlockers)
	w.blockLightUpdate = light.NewBlockLightUpdate(w.subChunkExplorer, w.lightFilters, w.lightEmitters)

	for _, blk := range block.GetRuntimeBlockStateRegistry().GetAllKnownStates() {
		w.registerTemplate(blk)
	}
	for _, blk := range knownBlocks {
		w.registerTemplate(blk)
	}
	return w
}

// SetLogger sets World::$logger.
func (w *World) SetLogger(logger log.Logger) { w.logger = logger }

// GetID is a port of World::getId.
func (w *World) GetID() int { return w.id }

// GetFolderName is a port of World::getFolderName.
func (w *World) GetFolderName() string { return w.folderName }

// GetDisplayName is a port of World::getDisplayName.
func (w *World) GetDisplayName() string { return w.displayName }

// SetDisplayName is a port of World::setDisplayName. The name is written to level.dat by
// WorldManager when the world is saved (WorldData::setName).
func (w *World) SetDisplayName(name string) {
	event.Call(worldevent.NewWorldDisplayNameChangeEvent(w, w.displayName, name))
	w.displayName = name
}

// GetChunk is a port of World::getChunk: a non-generating lookup (unlike GetOrLoadChunk/
// generateChunkOnly, this never creates a chunk that isn't already loaded) - the
// utils.ChunkSource capability SubChunkExplorer (and so the light engine) needs.
func (w *World) GetChunk(chunkX, chunkZ int) (*format.Chunk, bool) {
	c, ok := w.chunks[chunkKey(chunkX, chunkZ)]
	return c, ok
}

// registerTemplate records blk as the reconstruction template for its state ID, with the per-state
// light, blast resistance and random tick tables (RuntimeBlockStateRegistry's static arrays). New
// registers every state of RuntimeBlockStateRegistry; this also covers states that aren't in it
// (e.g. UnknownBlock).
func (w *World) registerTemplate(blk block.Behavior) {
	stateID := int32(blk.GetStateId())
	if _, ok := w.stateTemplates[stateID]; ok {
		return
	}
	w.stateTemplates[stateID] = blk.Clone()
	w.registryVersion++

	w.lightFilters[stateID] = min(15, blk.GetLightFilter()+light.BaseLightFilter)
	w.lightEmitters[stateID] = blk.GetLightLevel()
	w.directSkyLightBlockers[stateID] = blk.BlocksDirectSkyLight()
	w.blastResistance[stateID] = blk.GetBreakInfo().GetBlastResistance()
	if blk.TicksRandomly() {
		w.randomTickBlocks[stateID] = true
	}
}

func chunkKey(chunkX, chunkZ int) [2]int { return [2]int{chunkX, chunkZ} }

// GetOrLoadChunk generates, populates and lights the chunk right away if it isn't already,
// running population through a SyncGeneratorExecutor on the calling goroutine. It has no PHP
// counterpart (PHP always waits for requestChunkPopulation's promise): it's for startup and
// tests. Players get their chunks through RequestChunkPopulation, which doesn't block the tick.
func (w *World) GetOrLoadChunk(chunkX, chunkZ int) *format.Chunk {
	chunk := w.loadChunk(chunkX, chunkZ)
	if chunk == nil || !chunk.IsPopulated() {
		if w.IsChunkLocked(chunkX, chunkZ) {
			// An async population is in flight: finishing it synchronously breaks its locks, so
			// its result is discarded when it comes back.
			for xx := -1; xx <= 1; xx++ {
				for zz := -1; zz <= 1; zz++ {
					w.UnlockChunk(chunkX+xx, chunkZ+zz, nil)
				}
			}
		}
		var center *format.Chunk
		if chunk != nil {
			center = chunk.Clone()
		}
		adjacent := w.getAdjacentChunks(chunkX, chunkZ)
		for relative, c := range adjacent {
			if c != nil {
				adjacent[relative] = c.Clone()
			}
		}
		result := runPopulation(w.generatorSetupParameters(), w.generator, w.registrySnapshot(), chunkX, chunkZ, center, adjacent)
		for _, blk := range result.NewTemplates {
			w.registerTemplate(blk)
		}
		oldChunk := chunk
		w.SetChunk(chunkX, chunkZ, result.Center)
		for relative, c := range result.Adjacent {
			w.SetChunk(chunkX+relative[0], chunkZ+relative[1], c)
		}
		chunk = result.Center
		if oldChunk == nil || !oldChunk.IsPopulated() {
			if event.HasHandlers[worldevent.ChunkPopulateEvent]() {
				event.Call(worldevent.NewChunkPopulateEvent(w, chunkX, chunkZ, chunk))
			}
			for _, listener := range w.GetChunkListeners(chunkX, chunkZ) {
				listener.OnChunkPopulated(chunkX, chunkZ, chunk)
			}
		}
		if resolver, ok := w.chunkPopulationRequestMap[chunkKey(chunkX, chunkZ)]; ok {
			if _, active := w.activeChunkPopulationTasks[chunkKey(chunkX, chunkZ)]; !active {
				delete(w.chunkPopulationRequestMap, chunkKey(chunkX, chunkZ))
				resolver.Resolve(chunk)
			}
		}
	}
	if lit, known := chunk.IsLightPopulated(); !known || !lit {
		result := computeChunkLight(chunk.Clone(), w.registrySnapshot())
		chunk.SetHeightMapArray(result.HeightMap)
		for y, lightArray := range result.BlockLight {
			chunk.GetSubChunk(y).SetBlockLightArray(lightArray)
		}
		for y, lightArray := range result.SkyLight {
			chunk.GetSubChunk(y).SetBlockSkyLightArray(lightArray)
		}
		chunk.SetLightPopulated(true, true)
	}
	return chunk
}

// generateChunkOnly returns the loaded chunk, loading it from disk or generating it (without
// population) if needed. PHP's getBlockAt and friends read unloaded terrain as air instead; this
// port's block accessors generate on demand, which is cheap (see World.GetBlockAt's callers). A
// chunk that's locked for async population isn't generated here (nil is returned), so the main
// thread doesn't race the worker for it.
func (w *World) generateChunkOnly(chunkX, chunkZ int) *format.Chunk {
	if c := w.loadChunk(chunkX, chunkZ); c != nil {
		return c
	}
	if w.IsChunkLocked(chunkX, chunkZ) {
		return nil
	}
	c := w.generator.GenerateChunk(chunkX, chunkZ)
	w.SetChunk(chunkX, chunkZ, c)
	return c
}

// fireOnChunkLoaded is a port of the "$oldChunk === null" branch of setChunk's own ChunkListener
// dispatch (see World::setChunk) - this port's own chunk-creation path (generateChunkOnly) always
// corresponds to that branch, since nothing here ever replaces an already-loaded chunk wholesale
// (see ChunkListener.OnChunkChanged's own doc comment on that same gap).
func (w *World) fireOnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk) {
	for _, listener := range w.GetChunkListeners(chunkX, chunkZ) {
		listener.OnChunkLoaded(chunkX, chunkZ, chunk)
	}
}

// resolveBlockState is BaseWorldProvider's palette deserialization: a saved block state to its
// internal state ID through GlobalBlockStateHandlers' deserializer. States that can't be
// deserialized become GlobalBlockStateHandlers::getUnknownBlockStateData() (the "update!" block),
// like PHP (the block data upgrader for older saves isn't ported).
func (w *World) resolveBlockState(data bedrock.BlockStateData) (int32, bool) {
	deserializer := worldformatio.GetBlockStateDeserializer()
	stateID, err := deserializer.Deserialize(data)
	if err != nil {
		if w.logger != nil {
			w.logger.Debug(fmt.Sprintf("Unknown block state %s: %v", data.Name, err))
		}
		stateID, err = deserializer.Deserialize(worldformatio.GetUnknownBlockStateData())
		if err != nil {
			return 0, false
		}
	}
	return int32(stateID), true
}

// lookupBlockState is GlobalBlockStateHandlers::getSerializer()->serialize($stateId): the state data
// saved for an internal state ID.
func (w *World) lookupBlockState(stateID int32) (bedrock.BlockStateData, error) {
	return worldformatio.GetBlockStateSerializer().Serialize(int(stateID))
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
// no-op if no provider is open. It's the chunk half of World::save(true), which also fires
// WorldSaveEvent.
func (w *World) SaveAll() error {
	if w.provider == nil {
		return nil
	}
	event.Call(worldevent.NewWorldSaveEvent(w))
	for key, chunk := range w.chunks {
		if err := worldio.SaveChunk(w.provider, int32(key[0]), int32(key[1]), chunk, w.lookupBlockState); err != nil {
			return err
		}
		if err := worldio.SaveEntities(w.provider, int32(key[0]), int32(key[1]), w.saveChunkEntities(key[0], key[1])); err != nil {
			return err
		}
	}
	return nil
}

// IsClosed reports whether the world was unloaded (Close was called): a Position in it is no
// longer valid.
func (w *World) IsClosed() bool { return w.closed }

// Close saves every loaded chunk (see SaveAll) and closes the on-disk provider, if one is open.
func (w *World) Close() error {
	w.closed = true
	if w.provider == nil {
		return nil
	}
	if err := w.SaveAll(); err != nil {
		return err
	}
	return w.provider.Close()
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
	stateID := int32(block.VanillaAir().GetStateId())
	if chunk := w.generateChunkOnly(x>>4, z>>4); chunk != nil {
		stateID = chunk.GetBlockStateID(x&0xf, y, z&0xf)
	}
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

// SetBlock is a port of World::setBlockAt with $update always true (block.World has no caller that
// ever needs the $update=false fast path). Also registers blk as a state template (see
// registerTemplate) so it can be read back later even if it wasn't in New's knownBlocks list.
func (w *World) SetBlock(pos block.Position, blk block.Behavior) error {
	return w.SetBlockUpdate(pos, blk, true)
}

// SetBlockUpdate is World::setBlock with its $update parameter: when update is false, light isn't
// recalculated and neighbours aren't notified (Leaves and Farmland use this).
func (w *World) SetBlockUpdate(pos block.Position, blk block.Behavior, update bool) error {
	x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()
	w.registerTemplate(blk)
	chunk := w.generateChunkOnly(x>>4, z>>4)
	if chunk == nil {
		return fmt.Errorf("Cannot set a block in un-generated terrain")
	}
	// setBlockAt breaks any population lock: the async result is discarded and redone.
	w.UnlockChunk(x>>4, z>>4, nil)
	chunk.SetBlockStateID(x&0xf, y, z&0xf, int32(blk.GetStateId()))

	chunkPos := [2]int{x >> 4, z >> 4}

	if w.changedBlocks == nil {
		w.changedBlocks = map[[2]int]map[[3]int]math.Vector3{}
	}
	if w.changedBlocks[chunkPos] == nil {
		w.changedBlocks[chunkPos] = map[[3]int]math.Vector3{}
	}
	w.changedBlocks[chunkPos][[3]int{x, y, z}] = math.NewVector3(float64(x), float64(y), float64(z))

	for _, listener := range w.GetChunkListeners(x>>4, z>>4) {
		listener.OnBlockChanged(math.NewVector3(float64(x), float64(y), float64(z)))
	}

	if update {
		// Matches updateAllLight's own guard: recalculating light against a chunk whose light
		// hasn't been calculated at all yet would be meaningless (and RecalculateNode's BFS
		// assumes its starting light values are already meaningful).
		if lit, known := chunk.IsLightPopulated(); known && lit {
			w.skyLightUpdate.RecalculateNode(x, y, z)
			w.blockLightUpdate.RecalculateNode(x, y, z)
		}
		w.internalNotifyNeighbourBlockUpdate(x, y, z)
	}
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
	if chunk == nil {
		return nil, false
	}
	return chunk.GetTile(x&0xf, y, z&0xf)
}

// AddTile is a port of World::addTile.
func (w *World) AddTile(t block.Tile) {
	pos := t.GetPosition()
	chunk := w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4)
	if chunk == nil {
		panic(fmt.Sprintf("Attempted to create tile %T in unloaded chunk", t))
	}
	chunk.AddTile(t)
}

// RemoveTile is a port of World::removeTile - also satisfies tile.World's own RemoveTile (called
// by TileBase.Close() implementations, matching Tile::close()'s real `$this->lastUsedChunk?->
// removeTile($this)`-style bookkeeping).
func (w *World) RemoveTile(t block.Tile) {
	pos := t.GetPosition()
	if chunk := w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4); chunk != nil {
		chunk.RemoveTile(t)
	}
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

// GetOrLoadChunkAtPosition is a port of World::getOrLoadChunkAtPosition (generating the chunk if
// needed, see generateChunkOnly; never populating it).
func (w *World) GetOrLoadChunkAtPosition(pos block.Position) (block.Chunk, bool) {
	chunk := w.generateChunkOnly(pos.FloorX()>>4, pos.FloorZ()>>4)
	if chunk == nil {
		return nil, false
	}
	return chunkAdapter{chunk}, true
}

// viewer is the local surface AddSound/AddParticle/BroadcastPacketToViewers need to actually
// deliver a packet to a connected player - declared locally (matching this port's established
// forward-compatible-local-interface convention) rather than importing pocketmine/player.Player
// directly, which would create an import cycle (player already imports this package). Every
// ChunkListener registered in this port is in practice a *player.Player (see
// Player.RequestChunks), so filtering GetChunkListeners down to ones satisfying viewer is
// equivalent to real PHP's own separate playerChunkListeners side-table without needing one.
type viewer interface {
	SendPacket(pk packet.Packet)
}

// getViewersForPosition is a port of World::getViewersForPosition (via getChunkPlayers) - see
// viewer's own doc comment on why this filters GetChunkListeners instead of maintaining a separate
// playerChunkListeners table.
func (w *World) getViewersForPosition(pos math.Vector3) []viewer {
	listeners := w.GetChunkListeners(pos.FloorX()>>4, pos.FloorZ()>>4)
	viewers := make([]viewer, 0, len(listeners))
	for _, l := range listeners {
		if v, ok := l.(viewer); ok {
			viewers = append(viewers, v)
		}
	}
	return viewers
}

// BroadcastPacketToViewers is a port of World::broadcastPacketToViewers (minus the packet-buffering
// optimisation real PHP does via packetBuffersByChunk/flushed at the end of the tick - this port
// sends immediately, a purely internal timing difference invisible to the client).
func (w *World) BroadcastPacketToViewers(pos math.Vector3, pk packet.Packet) {
	for _, v := range w.getViewersForPosition(pos) {
		v.SendPacket(pk)
	}
}

// AddSound is a port of World::addSound for the default recipients (every player viewing pos).
func (w *World) AddSound(pos math.Vector3, s sound.Sound) { w.AddSoundFor(pos, s, nil) }

// AddSoundFor is a port of World::addSound: players nil means everyone viewing pos; otherwise only
// those of players who are viewing pos hear it.
func (w *World) AddSoundFor(pos math.Vector3, s sound.Sound, players []EntityViewer) {
	defaultRecipients := players == nil
	if event.HasHandlers[worldevent.WorldSoundEvent]() {
		ev := worldevent.NewWorldSoundEvent(w, s, pos, w.recipientsForEvent(pos, players))
		event.Call(ev)
		if ev.IsCancelled() {
			return
		}
		s = ev.GetSound()
		players, defaultRecipients = recipientsFromEvent(ev.GetRecipients()), false
	}
	w.broadcastEncoded(pos, s.Encode(pos, w.translator), players, defaultRecipients)
}

// AddParticle is a port of World::addParticle for the default recipients.
func (w *World) AddParticle(pos math.Vector3, p particle.Particle) { w.AddParticleFor(pos, p, nil) }

// AddParticleFor is a port of World::addParticle with explicit recipients (see AddSoundFor).
func (w *World) AddParticleFor(pos math.Vector3, p particle.Particle, players []EntityViewer) {
	defaultRecipients := players == nil
	if event.HasHandlers[worldevent.WorldParticleEvent]() {
		ev := worldevent.NewWorldParticleEvent(w, p, pos, w.recipientsForEvent(pos, players))
		event.Call(ev)
		if ev.IsCancelled() {
			return
		}
		p = ev.GetParticle()
		players, defaultRecipients = recipientsFromEvent(ev.GetRecipients()), false
	}
	w.broadcastEncoded(pos, p.Encode(pos, w.translator), players, defaultRecipients)
}

// recipientsForEvent is the `$players ??= $this->getViewersForPosition($pos)` of
// addSound/addParticle, as the event's recipient list.
func (w *World) recipientsForEvent(pos math.Vector3, players []EntityViewer) []any {
	var result []any
	if players == nil {
		for _, v := range w.getViewersForPosition(pos) {
			result = append(result, v)
		}
		return result
	}
	for _, p := range players {
		result = append(result, p)
	}
	return result
}

func recipientsFromEvent(recipients []any) []EntityViewer {
	result := make([]EntityViewer, 0, len(recipients))
	for _, r := range recipients {
		if v, ok := r.(EntityViewer); ok {
			result = append(result, v)
		}
	}
	return result
}

// broadcastEncoded sends the encoded packets of a sound or particle: to every viewer of pos, or
// only to those of players who view pos (World::filterViewersForPosition).
func (w *World) broadcastEncoded(pos math.Vector3, packets []packet.Packet, players []EntityViewer, defaultRecipients bool) {
	if len(packets) == 0 {
		return
	}
	if defaultRecipients {
		for _, pk := range packets {
			w.BroadcastPacketToViewers(pos, pk)
		}
		return
	}
	candidates := w.getViewersForPosition(pos)
	for _, p := range players {
		for _, c := range candidates {
			if c == viewer(p) {
				for _, pk := range packets {
					p.SendPacket(pk)
				}
				break
			}
		}
	}
}

// scheduledBlockUpdate is one entry in World's scheduled-block-update queue (see
// ScheduleDelayedBlockUpdate/updateScheduledBlocks in tick.go).
type scheduledBlockUpdate struct {
	pos     [3]int
	dueTick int64
}

// ScheduleDelayedBlockUpdate is a port of World::scheduleDelayedBlockUpdate: the position's block
// gets OnScheduledUpdate() called on it once delay ticks have passed (see updateScheduledBlocks,
// called from DoTick). Matches the PHP original's dedup rule exactly: if a scheduled update for
// this exact position already exists with an equal-or-shorter delay, this new request is dropped
// (the existing one will already fire in time).
func (w *World) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {
	x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()
	if !w.IsInWorld(x, y, z) {
		return
	}
	key := [3]int{x, y, z}
	if existingDelay, ok := w.scheduledUpdateDelay[key]; ok && existingDelay <= delay {
		return
	}
	w.scheduledUpdateDelay[key] = delay
	w.scheduledUpdates = append(w.scheduledUpdates, scheduledBlockUpdate{pos: key, dueTick: w.currentTick + int64(delay)})
}

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

// GetSkyLightReduction is a port of World::getSkyLightReduction: how many points of sky light are
// currently subtracted for time of day (weather/rain isn't factored in - matches real PHP's own
// "TODO: check rain and thunder level" on computeSkyLightReduction). Recomputed once per tick in
// DoTick, not lazily, matching the real field's own update point.
func (w *World) GetSkyLightReduction() int { return w.skyLightReduction }

// GetSunAnglePercentage is a port of World::getSunAnglePercentage - the percentage of a full circle
// away from noon the sun currently is (see computeSunAnglePercentage in tick.go for the formula).
// Recomputed once per tick in DoTick from World.time.
func (w *World) GetSunAnglePercentage() float64 { return w.sunAnglePercentage }

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

// fullCubeChecker is the local surface GetSafeSpawn's flatness checks need - IsFullCube is
// promoted from *block.Block, not part of block.Behavior itself (same "declare the exact promoted
// method this file needs" convention as positionable above).
type fullCubeChecker interface {
	IsFullCube() bool
}

func (w *World) isFullCube(x, y, z int) bool {
	fc, ok := w.GetBlockAt(x, y, z).(fullCubeChecker)
	return ok && fc.IsFullCube()
}

// GetSafeSpawn is a port of World::getSafeSpawn: searches vertically at spawn's x/z column for a
// safe (2-block-high air pocket resting on a solid full cube) position near spawn's own height,
// falling back to wherever the search ends up if nothing better is found. Doesn't accept spawn's
// own nil-Vector3 default (real PHP's `?Vector3 $spawn = null` falling back to
// `$this->getSpawnLocation()`) - callers wanting that exact default should just pass
// w.GetSpawnLocation() themselves.
func (w *World) GetSafeSpawn(spawn math.Vector3) math.Vector3 {
	v := spawn.Floor()
	x, z := int(v.X), int(v.Z)

	y := int(v.Y)
	if YMax-2 < y {
		y = YMax - 2
	}

	wasAir := w.GetBlockAt(x, y-1, z).GetTypeId() == block.AIR
	for ; y > YMin; y-- {
		if w.isFullCube(x, y, z) {
			if wasAir {
				y++
			}
			break
		}
		wasAir = true
	}

	for ; y >= YMin && y < YMax; y++ {
		if !w.isFullCube(x, y+1, z) {
			if !w.isFullCube(x, y, z) {
				resultY := float64(y)
				if y == int(spawn.Y) {
					resultY = spawn.Y
				}
				return math.NewVector3(spawn.X, resultY, spawn.Z)
			}
		} else {
			y++
		}
	}

	return math.NewVector3(spawn.X, float64(y), spawn.Z)
}

// GetSpawnLocation is a port of World::getSpawnLocation - simplified to a bare Vector3, not a
// Position (block.Position always needs a World reference to construct, which callers already
// have here - see e.g. UseBreakOn's identical reasoning for using math.Vector3 directly).
func (w *World) GetSpawnLocation() math.Vector3 { return w.spawnLocation }

// SetSpawnLocation is a port of World::setSpawnLocation: fires SpawnChangeEvent and syncs the new
// spawn point to every player in the world (NetworkSession::syncWorldSpawnPoint).
func (w *World) SetSpawnLocation(pos math.Vector3) {
	previousSpawn := w.spawnLocation
	w.spawnLocation = pos
	event.Call(worldevent.NewSpawnChangeEvent(w, entityevent.Position{Vector3: previousSpawn, World: w}))

	pk := worldSpawnPacket(pos)
	for _, p := range w.GetPlayers() {
		p.SendPacket(pk)
	}
}

// worldSpawnPacket is SetSpawnPositionPacket::worldSpawn (NetworkSession::syncWorldSpawnPoint).
func worldSpawnPacket(pos math.Vector3) packet.Packet {
	const int32Min = -2147483648
	return &packet.SetSpawnPosition{
		SpawnType:     packet.SpawnTypeWorld,
		Position:      protocol.BlockPos{int32(pos.FloorX()), int32(pos.FloorY()), int32(pos.FloorZ())},
		Dimension:     packet.DimensionOverworld,
		SpawnPosition: protocol.BlockPos{int32Min, int32Min, int32Min},
	}
}

// GetPlayers is a port of World::getPlayers: the players in this world (the entities that are
// packet viewers).
func (w *World) GetPlayers() []EntityViewer {
	var players []EntityViewer
	for _, id := range w.sortedEntityIDs() {
		if v, ok := w.entities[id].(EntityViewer); ok {
			players = append(players, v)
		}
	}
	return players
}

// sortedEntityIDs is the entity table's keys in ascending order (PHP arrays keep insertion order,
// and runtime IDs only ever increase).
func (w *World) sortedEntityIDs() []int {
	ids := make([]int, 0, len(w.entities))
	for id := range w.entities {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// IsSpawnChunk is a port of World::isSpawnChunk.
func (w *World) IsSpawnChunk(chunkX, chunkZ int) bool {
	spawnChunkX, spawnChunkZ := int(w.spawnLocation.X)>>4, int(w.spawnLocation.Z)>>4
	dx, dz := chunkX-spawnChunkX, chunkZ-spawnChunkZ
	if dx < 0 {
		dx = -dx
	}
	if dz < 0 {
		dz = -dz
	}
	return dx <= 1 && dz <= 1
}

// GetBiomeID is a port of World::getBiomeId. This port's chunks are always generatable on demand
// (see GetBlockAt's identical reasoning), so unlike real PHP this never falls back to a bare
// BiomeIds::OCEAN guess for ungenerated terrain - it just generates it, same as everywhere else in
// this port.
func (w *World) GetBiomeID(x, y, z int) int32 {
	chunk := w.generateChunkOnly(x>>4, z>>4)
	if chunk == nil {
		return 0 // BiomeIds::OCEAN, what PHP returns for ungenerated terrain
	}
	return chunk.GetBiomeID(x&0xf, y, z&0xf)
}

// GetBiome is a port of World::getBiome.
func (w *World) GetBiome(x, y, z int) *biome.Biome {
	return w.biomeRegistry.GetBiome(int(w.GetBiomeID(x, y, z)))
}

// SetBiomeID is a port of World::setBiomeId.
func (w *World) SetBiomeID(x, y, z int, biomeID int32) {
	chunk := w.generateChunkOnly(x>>4, z>>4)
	if chunk == nil {
		return
	}
	w.UnlockChunk(x>>4, z>>4, nil)
	chunk.SetBiomeID(x&0xf, y, z&0xf, biomeID)
}

// IsChunkLoaded is a port of World::isChunkLoaded.
func (w *World) IsChunkLoaded(chunkX, chunkZ int) bool {
	_, ok := w.chunks[chunkKey(chunkX, chunkZ)]
	return ok
}

// IsChunkGenerated is a port of World::isChunkGenerated. Real PHP's own version actually generates
// the chunk as a side effect of checking (it calls loadChunk, which creates the chunk if it isn't
// loaded already) - this port's chunks are always generatable on demand regardless of this check's
// answer, so doing the same here would make this method trivially always return true. Matching
// IsChunkLoaded instead keeps this a real, side-effect-free query.
func (w *World) IsChunkGenerated(chunkX, chunkZ int) bool { return w.IsChunkLoaded(chunkX, chunkZ) }

// IsChunkPopulated is a port of World::isChunkPopulated.
func (w *World) IsChunkPopulated(chunkX, chunkZ int) bool {
	chunk, ok := w.chunks[chunkKey(chunkX, chunkZ)]
	return ok && chunk.IsPopulated()
}

// GetLoadedChunks is a port of World::getLoadedChunks.
func (w *World) GetLoadedChunks() map[[2]int]*format.Chunk { return w.chunks }

// GetTickRateTime is a port of World::getTickRateTime: how long the last tick took, in ms.
func (w *World) GetTickRateTime() float64 { return w.tickRateTime }

// GetTickingChunks is a port of World::getTickingChunks: the chunks registered for ticking.
func (w *World) GetTickingChunks() [][2]int {
	chunks := make([][2]int, 0, len(w.tickingChunks))
	for key := range w.tickingChunks {
		chunks = append(chunks, key)
	}
	return chunks
}
