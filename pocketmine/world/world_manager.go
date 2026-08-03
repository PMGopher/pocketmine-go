package world

import (
	"fmt"
	"path/filepath"
	"strings"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	worldio "pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/generator"
)

// ticksPerAutoSave mirrors WorldManager::TICKS_PER_AUTOSAVE (300 * Server::TARGET_TICKS_PER_SECOND,
// and this port's server is likewise fixed at 20 TPS - see cmd/pocketmine-go's own runTickLoop).
const ticksPerAutoSave = 300 * 20

// WorldManager is a port of pocketmine\world\WorldManager: the multi-world container that owns
// every currently-loaded World, tracks which one is the default, and drives all of their ticks and
// autosave together. Not ported: anything touching Player (unloadWorld's player eviction/
// teleport, doAutoSave's per-player save) or the plugin event bus (WorldInitEvent/WorldLoadEvent/
// WorldUnloadEvent) - this port has no Player type reachable from the world package and no
// event-bus-that-can-cancel-a-world-unload wired up yet, both documented gaps matching this port's
// established "no player/event infrastructure here yet" pattern elsewhere (see Explosion's own doc
// comment for the same category of gap).
type WorldManager struct {
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

// UnloadWorld is a port of WorldManager::unloadWorld. Not ported: player eviction/teleport and the
// cancellable WorldUnloadEvent (see WorldManager's own doc comment on why).
func (m *WorldManager) UnloadWorld(w *World, forceUnload bool) (bool, error) {
	if w == m.defaultWorld && !forceUnload {
		return false, fmt.Errorf("world manager: the default world cannot be unloaded while running, please switch worlds")
	}
	if w.IsDoingTick() {
		return false, fmt.Errorf("world manager: cannot unload a world during its own tick")
	}

	if w == m.defaultWorld {
		m.defaultWorld = nil
	}
	delete(m.worlds, w.GetID())
	delete(m.worldData, w.GetID())

	if err := w.Close(); err != nil {
		return false, fmt.Errorf("world manager: closing world %q: %w", w.GetFolderName(), err)
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

	m.nextID++
	w.id = m.nextID
	w.folderName = name
	w.displayName = wd.GetName()

	m.worlds[w.id] = w
	m.worldData[w.id] = wd
	return w, nil
}

// GenerateWorld is a port of WorldManager::generateWorld, minus background spawn-chunk
// pregeneration (real PHP's own $backgroundGeneration - a pure startup-latency optimisation this
// port's synchronous generation pipeline has no equivalent async task queue to run it on anyway;
// see ensurePopulated's own doc comment on why this port generates everything inline instead) and
// the WorldInitEvent/WorldLoadEvent (see WorldManager's own doc comment).
//
// gen must already be fully constructed by the caller (see generator.Factory's own doc comment on
// why - not every Generator this port has can be built from just a name + options string yet);
// generatorName/generatorOptions are recorded into level.dat purely so a later LoadWorld (in this
// process or a future run) can reconstruct an equivalent generator via generator.GetFactory.
func (m *WorldManager) GenerateWorld(name string, gen generator.Generator, seed int64, generatorName, generatorOptions string) (*World, error) {
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

	spawn := w.GetSafeSpawn(math.Vector3Zero())

	wd, err := worldio.GenerateWorldData(path, name, seed, worldio.GeneratorInfinite, generatorName, generatorOptions, spawn)
	if err != nil {
		return nil, fmt.Errorf("world manager: writing %q's level.dat: %w", name, err)
	}

	m.nextID++
	w.id = m.nextID
	w.folderName = name
	w.displayName = name

	m.worlds[w.id] = w
	m.worldData[w.id] = wd
	return w, nil
}

// FindEntity is a port of WorldManager::findEntity.
func (m *WorldManager) FindEntity(entityID int) (block.Entity, bool) {
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
		w.DoTick(currentTick)
	}

	if m.autoSave {
		m.autoSaveTicker++
		if m.autoSaveTicker >= m.autoSaveTicks {
			m.autoSaveTicker = 0
			m.doAutoSave()
		}
	}
}

// doAutoSave is a port of WorldManager::doAutoSave, minus per-player saving (no Player type - see
// WorldManager's own doc comment).
func (m *WorldManager) doAutoSave() {
	for id, w := range m.worlds {
		_ = w.SaveAll()
		if wd, ok := m.worldData[id]; ok {
			wd.SetTime(w.GetTime())
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
