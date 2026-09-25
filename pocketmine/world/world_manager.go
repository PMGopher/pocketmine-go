package world

import (
	"fmt"
	stdmath "math"
	"path/filepath"
	"pocketmine-go/pocketmine/event"
	worldevent "pocketmine-go/pocketmine/event/world"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/scheduler"
	"strings"
	"time"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	worldio "pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/generator"
)

// ticksPerAutoSave mirrors WorldManager::TICKS_PER_AUTOSAVE (300 * Server::TARGET_TICKS_PER_SECOND,
// and this port's server is likewise fixed at 20 TPS - see cmd/pocketmine-go's own runTickLoop).
const ticksPerAutoSave = 300 * 20

// WorldManager is a port of pocketmine\world\WorldManager: the multi-world container that owns
// every currently-loaded World, tracks which one is the default, and drives all of their ticks and
// autosave together.
type WorldManager struct {
	logger log.Logger

	dataPath    string
	translator  *convert.BlockTranslator
	knownBlocks []block.Behavior

	worlds       map[int]*World
	worldData    map[int]*worldio.WorldData
	defaultWorld *World
	nextID       int

	autoSave       bool
	autoSaveTicks  int64
	autoSaveTicker int64

	// asyncPool is Server::getAsyncPool(), which worlds populate chunks and calculate light on;
	// nil makes worlds do both synchronously.
	asyncPool *scheduler.AsyncPool
	// populationQueueSize is pocketmine.yml's chunk-generation.population-queue-size.
	populationQueueSize int
}

// SetAsyncPool sets the pool worlds loaded or generated from now on populate chunks on.
func (m *WorldManager) SetAsyncPool(pool *scheduler.AsyncPool) { m.asyncPool = pool }

// SetPopulationQueueSize sets pocketmine.yml's chunk-generation.population-queue-size for worlds
// loaded or generated from now on.
func (m *WorldManager) SetPopulationQueueSize(n int) { m.populationQueueSize = n }

// fastGenerators are the generators GeneratorManager registers with $fast = true: they're cheap
// enough to run on the main thread (World::__construct uses a SyncGeneratorExecutor for them).
var fastGenerators = map[string]bool{"flat": true}

// setUpGeneration is the generator part of World::__construct: an AsyncGeneratorExecutor whose
// workers build their own generator from the world's generator name, seed and options.
func (m *WorldManager) setUpGeneration(w *World, generatorName string, seed int64, generatorOptions string) {
	if m.logger != nil {
		w.SetLogger(log.NewPrefixedLogger(m.logger, "World: "+w.GetFolderName()))
	}
	if m.populationQueueSize > 0 {
		w.SetMaxConcurrentChunkPopulationTasks(m.populationQueueSize)
	}
	if m.asyncPool == nil {
		return
	}
	factory, ok := generator.GetFactory(strings.ToLower(generatorName))
	if !ok || fastGenerators[strings.ToLower(generatorName)] {
		w.SetGeneratorExecutor(nil, m.asyncPool)
		return
	}
	logger := m.logger
	if logger == nil {
		logger = log.Global()
	}
	w.SetGeneratorExecutor(NewAsyncGeneratorExecutor(logger, m.asyncPool, GeneratorExecutorSetupParameters{
		WorldMinY:    YMin,
		WorldMaxY:    YMax,
		NewGenerator: func() (generator.Generator, error) { return factory(seed, generatorOptions) },
	}), m.asyncPool)
}

// NewWorldManager is a port of WorldManager::__construct. translator/knownBlocks are shared by
// every World this manager ever loads or generates - matching real PHP's own global block-registry
// singletons (RuntimeBlockStateRegistry, GlobalBlockStateHandlers), which this port instead passes
// explicitly (see World.New's own doc comment on why).
func NewWorldManager(dataPath string, translator *convert.BlockTranslator, knownBlocks []block.Behavior) *WorldManager {
	return &WorldManager{
		dataPath:      dataPath,
		translator:    translator,
		knownBlocks:   knownBlocks,
		worlds:        map[int]*World{},
		worldData:     map[int]*worldio.WorldData{},
		autoSave:      true,
		autoSaveTicks: ticksPerAutoSave,
	}
}

// SetLogger sets the logger unload messages go to (PHP uses the server's logger).
func (m *WorldManager) SetLogger(logger log.Logger) { m.logger = logger }

// managedPlayer is what WorldManager needs from pocketmine\player\Player: unloadWorld moves
// players out of the world, and doAutoSave saves them.
type managedPlayer interface {
	EntityViewer
	TeleportTo(pos math.Vector3, w *World, yaw, pitch *float64) bool
	// Disconnect is Player::disconnect(reason, quitMessage, disconnectScreenMessage); nil means
	// the default.
	Disconnect(reason, quitMessage, disconnectScreenMessage any)
	IsSpawned() bool
	Save()
}

// GetWorlds is a port of WorldManager::getWorlds.
func (m *WorldManager) GetWorlds() map[int]*World { return m.worlds }

// GetDefaultWorld is a port of WorldManager::getDefaultWorld.
func (m *WorldManager) GetDefaultWorld() *World { return m.defaultWorld }

// SetDefaultWorld is a port of WorldManager::setDefaultWorld.
func (m *WorldManager) SetDefaultWorld(w *World) {
	if w == nil || (m.isWorldLoadedInstance(w) && w != m.defaultWorld) {
		m.defaultWorld = w
	}
}

func (m *WorldManager) isWorldLoadedInstance(w *World) bool {
	loaded, ok := m.GetWorldByName(w.GetFolderName())
	return ok && loaded == w
}

// IsWorldLoaded is a port of WorldManager::isWorldLoaded.
func (m *WorldManager) IsWorldLoaded(name string) bool {
	_, ok := m.GetWorldByName(name)
	return ok
}

// GetWorld is a port of WorldManager::getWorld.
func (m *WorldManager) GetWorld(worldID int) (*World, bool) {
	w, ok := m.worlds[worldID]
	return w, ok
}

// GetWorldByName is a port of WorldManager::getWorldByName - matches based on the folder name, not
// the display name, exactly like the real method's own doc comment warns.
func (m *WorldManager) GetWorldByName(name string) (*World, bool) {
	for _, w := range m.worlds {
		if w.GetFolderName() == name {
			return w, true
		}
	}
	return nil, false
}

func (m *WorldManager) worldPath(name string) string { return filepath.Join(m.dataPath, name) }

// IsWorldGenerated is a port of WorldManager::isWorldGenerated - minus the WorldProviderManager
// format-matching probe (this port only ever writes LevelDB worlds - see io/leveldb's own doc
// comment - so "generated" just means "this world's directory has a level.dat in it").
func (m *WorldManager) IsWorldGenerated(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	if _, ok := m.GetWorldByName(name); ok {
		return true
	}
	_, err := worldio.LoadWorldData(m.worldPath(name))
	return err == nil
}

// UnloadWorld is a port of WorldManager::unloadWorld.
func (m *WorldManager) UnloadWorld(w *World, forceUnload bool) (bool, error) {
	if w == m.defaultWorld && !forceUnload {
		return false, fmt.Errorf("world manager: the default world cannot be unloaded while running, please switch worlds")
	}
	if w.IsDoingTick() {
		return false, fmt.Errorf("world manager: cannot unload a world during its own tick")
	}

	ev := worldevent.NewWorldUnloadEvent(w)
	event.Call(ev)
	if !forceUnload && ev.IsCancelled() {
		return false, nil
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Unloading world \"%s\"", w.GetDisplayName())) // pocketmine.level.unloading
	}
	if players := w.GetPlayers(); len(players) != 0 {
		var safeSpawn *math.Vector3
		if m.defaultWorld != nil && m.defaultWorld != w {
			spawn := m.defaultWorld.GetSafeSpawn(m.defaultWorld.GetSpawnLocation())
			safeSpawn = &spawn
		}
		for _, v := range players {
			p, ok := v.(managedPlayer)
			if !ok {
				continue
			}
			if safeSpawn == nil {
				p.Disconnect("Forced default world unload", nil, nil)
			} else {
				p.TeleportTo(*safeSpawn, m.defaultWorld, nil, nil)
			}
		}
	}

	if w == m.defaultWorld {
		m.defaultWorld = nil
	}
	// World::onUnload -> World::save: level.dat is saved along with the chunks (Close).
	if wd, ok := m.worldData[w.GetID()]; ok {
		m.syncWorldData(w, wd)
		if err := wd.Save(m.worldPath(w.GetFolderName())); err != nil {
			return false, fmt.Errorf("world manager: saving %q's level.dat: %w", w.GetFolderName(), err)
		}
	}
	delete(m.worlds, w.GetID())
	delete(m.worldData, w.GetID())

	if err := w.Close(); err != nil {
		return false, fmt.Errorf("world manager: closing world %q: %w", w.GetFolderName(), err)
	}
	if w.generatorExecutor != nil {
		w.generatorExecutor.Shutdown()
	}
	return true, nil
}

// LoadWorld is a port of WorldManager::loadWorld. Not ported: format auto-upgrade (autoUpgrade's
// FormatConverter path - this port only ever reads/writes its own single LevelDB format, so there
// is no other format to convert from) and the WorldLoadEvent (see WorldManager's own doc comment).
func (m *WorldManager) LoadWorld(name string) (*World, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("world manager: invalid empty world name")
	}
	if existing, ok := m.GetWorldByName(name); ok {
		return existing, nil
	}
	if !m.IsWorldGenerated(name) {
		return nil, fmt.Errorf("world manager: world %q has not been generated", name)
	}

	path := m.worldPath(name)
	wd, err := worldio.LoadWorldData(path)
	if err != nil {
		return nil, fmt.Errorf("world manager: loading %q's level.dat: %w", name, err)
	}

	factory, ok := generator.GetFactory(wd.GetGenerator())
	if !ok {
		return nil, &generator.UnknownGeneratorError{Name: wd.GetGenerator()}
	}
	gen, err := factory(wd.GetSeed(), wd.GetGeneratorOptions())
	if err != nil {
		return nil, fmt.Errorf("world manager: constructing %q's generator: %w", name, err)
	}

	w := New(gen, m.translator, m.knownBlocks)
	if err := w.OpenProvider(path); err != nil {
		return nil, fmt.Errorf("world manager: opening %q's world data: %w", name, err)
	}
	w.SetTime(wd.GetTime())
	w.spawnLocation = wd.GetSpawn()
	w.difficulty = wd.GetDifficulty()

	m.nextID++
	w.id = m.nextID
	w.folderName = name
	w.displayName = wd.GetName()
	m.setUpGeneration(w, wd.GetGenerator(), wd.GetSeed(), wd.GetGeneratorOptions())

	m.worlds[w.id] = w
	m.worldData[w.id] = wd

	event.Call(worldevent.NewWorldLoadEvent(w))
	return w, nil
}

// GenerateWorld is a port of WorldManager::generateWorld, minus background spawn-chunk
// pregeneration, which the server does (Server.generateSpawnTerrain: it needs ChunkSelector from
// the player package).
//
// gen must already be fully constructed by the caller (see generator.Factory's own doc comment on
// why - not every Generator this port has can be built from just a name + options string yet).
// options.SpawnPosition is written into level.dat exactly as given, unadjusted - matching real
// PHP's own generateWorld/BedrockWorldData::generate, which never runs a safe-spawn search at
// generation time either (see World.GetSafeSpawn's own doc comment - that search happens lazily,
// wherever a caller actually needs a safe point to put something, not baked into world creation).
func (m *WorldManager) GenerateWorld(name string, gen generator.Generator, options *WorldCreationOptions) (*World, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("world manager: invalid empty world name")
	}
	if m.IsWorldGenerated(name) {
		return nil, fmt.Errorf("world manager: world %q already exists", name)
	}

	path := m.worldPath(name)
	w := New(gen, m.translator, m.knownBlocks)
	if err := w.OpenProvider(path); err != nil {
		return nil, fmt.Errorf("world manager: opening %q's world data: %w", name, err)
	}

	wd, err := worldio.GenerateWorldData(path, name, options.Seed, worldio.GeneratorInfinite, options.GeneratorName, options.GeneratorOptions, options.SpawnPosition)
	if err != nil {
		return nil, fmt.Errorf("world manager: writing %q's level.dat: %w", name, err)
	}
	wd.SetDifficulty(options.Difficulty)
	if err := wd.Save(path); err != nil {
		return nil, fmt.Errorf("world manager: writing %q's level.dat: %w", name, err)
	}
	w.spawnLocation = options.SpawnPosition
	w.difficulty = options.Difficulty

	m.nextID++
	w.id = m.nextID
	w.folderName = name
	w.displayName = name
	m.setUpGeneration(w, options.GeneratorName, options.Seed, options.GeneratorOptions)

	m.worlds[w.id] = w
	m.worldData[w.id] = wd

	event.Call(worldevent.NewWorldInitEvent(w))
	event.Call(worldevent.NewWorldLoadEvent(w))
	return w, nil
}

// FindEntity is a port of WorldManager::findEntity.
func (m *WorldManager) FindEntity(entityID int) (Entity, bool) {
	for _, w := range m.worlds {
		if e, ok := w.GetEntity(entityID); ok {
			return e, true
		}
	}
	return nil, false
}

// Tick is a port of WorldManager::tick, minus its own per-world tickRateTime/slow-tick-warning
// timing instrumentation (a pure diagnostics concern, not behaviour).
func (m *WorldManager) Tick(currentTick int64) {
	for _, w := range m.worlds {
		worldTime := time.Now()
		w.DoTick(currentTick)
		tickMs := float64(time.Since(worldTime).Microseconds()) / 1000
		w.tickRateTime = tickMs
		if tickMs >= 50 && m.logger != nil {
			m.logger.Debug(fmt.Sprintf("[World: %s] Tick took too long: %gms (%g ticks)", w.GetFolderName(), tickMs, stdmath.Round(tickMs/50*100)/100))
		}
	}

	if m.autoSave {
		m.autoSaveTicker++
		if m.autoSaveTicker >= m.autoSaveTicks {
			m.autoSaveTicker = 0
			m.doAutoSave()
		}
	}
}

// doAutoSave is a port of WorldManager::doAutoSave.
func (m *WorldManager) doAutoSave() {
	for id, w := range m.worlds {
		for _, v := range w.GetPlayers() {
			if p, ok := v.(managedPlayer); ok && p.IsSpawned() {
				p.Save()
			}
		}
		_ = w.SaveAll()
		if wd, ok := m.worldData[id]; ok {
			m.syncWorldData(w, wd)
			_ = wd.Save(m.worldPath(w.GetFolderName()))
		}
	}
}

// GetAutoSave is a port of WorldManager::getAutoSave.
func (m *WorldManager) GetAutoSave() bool { return m.autoSave }

// SetAutoSave is a port of WorldManager::setAutoSave.
func (m *WorldManager) SetAutoSave(value bool) { m.autoSave = value }

// GetAutoSaveInterval is a port of WorldManager::getAutoSaveInterval.
func (m *WorldManager) GetAutoSaveInterval() int64 { return m.autoSaveTicks }

// SetAutoSaveInterval is a port of WorldManager::setAutoSaveInterval.
func (m *WorldManager) SetAutoSaveInterval(autoSaveTicks int64) error {
	if autoSaveTicks <= 0 {
		return fmt.Errorf("world manager: autosave ticks must be positive")
	}
	m.autoSaveTicks = autoSaveTicks
	return nil
}

// syncWorldData copies the World state PHP keeps directly in its WorldData (time, spawn, display
// name, difficulty) into wd before it's saved.
func (m *WorldManager) syncWorldData(w *World, wd *worldio.WorldData) {
	wd.SetTime(w.GetTime())
	wd.SetSpawn(w.GetSpawnLocation())
	wd.SetName(w.GetDisplayName())
	wd.SetDifficulty(w.GetDifficulty())
}

// GetSeed is World::getSeed: the seed in the world's level.dat (this port's World doesn't own its
// WorldData, see syncWorldData).
func (m *WorldManager) GetSeed(w *World) int64 {
	if wd, ok := m.worldData[w.GetID()]; ok {
		return wd.GetSeed()
	}
	return 0
}

// SaveWorld is World::save(true) for a loaded world: its chunks, entities and level.dat.
func (m *WorldManager) SaveWorld(w *World) error {
	if err := w.SaveAll(); err != nil {
		return err
	}
	if wd, ok := m.worldData[w.GetID()]; ok {
		m.syncWorldData(w, wd)
		return wd.Save(m.worldPath(w.GetFolderName()))
	}
	return nil
}
