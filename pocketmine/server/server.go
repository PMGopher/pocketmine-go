package server

import (
	_ "embed"
	"fmt"
	"log/slog"
	stdmath "math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/console"
	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	serverevent "pocketmine-go/pocketmine/event/server"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/network/mcpe/raklib"
	"pocketmine-go/pocketmine/network/query"
	"pocketmine-go/pocketmine/network/upnp"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/resourcepacks"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/timings"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator"

	// Linked for their init(): EntityFactory registrations and the world/block hooks that create
	// item entities, experience orbs, falling blocks and primed TNT.
	_ "pocketmine-go/pocketmine/entity/object"
	_ "pocketmine-go/pocketmine/entity/projectile"
	// Linked for its init(), which gives NetworkSession its packet handlers.
	_ "pocketmine-go/pocketmine/network/mcpe/handler"
)

// Server constants, Server::* .
const (
	BroadcastChannelAdministrative = command.BroadcastChannelAdministrative
	BroadcastChannelUsers          = player.BroadcastChannelUsers

	DefaultServerName      = pocketmine.Name + " Server"
	DefaultMaxPlayers      = 20
	DefaultPortIPv4        = 19132
	DefaultPortIPv6        = 19133
	DefaultMaxViewDistance = 16

	// TargetTicksPerSecond is the tick rate the server tries to keep up with.
	TargetTicksPerSecond = 20
	// TargetSecondsPerTick is the time in seconds that the server tries to take per tick.
	TargetSecondsPerTick = 1.0 / TargetTicksPerSecond

	ticksPerWorldCacheClear       = 5 * TargetTicksPerSecond
	ticksPerTPSOverloadWarning    = 5 * TargetTicksPerSecond
	tpsOverloadWarningThreshold   = TargetTicksPerSecond * 0.6
	defaultAsyncCompressionThresh = 10_000
)

//go:embed resources/pocketmine.yml
var defaultPocketmineYml []byte

func init() {
	PruneChunkCachesFunc = mcpe.PruneChunkCaches
}

// Server is a port of pocketmine\Server.
//
// Not ported (see AGENTS.md): plugins (PluginManager: the design is undecided, so there are no
// plugin schedulers to tick and no plugin enable phases), the update checker,
// anonymous usage statistics (SendUsageTask), crash dumps, the signal handler (main.go handles
// SIGINT/SIGTERM), compression settings (gophertunnel compresses; network.batch-threshold is
// passed on as its compression threshold) and the AuthKeyProvider (gophertunnel verifies logins).
type Server struct {
	// mu serialises the tick with packet handling, which PHP does on one thread (see
	// mcpe.Server's Locker).
	mu sync.Mutex

	logger     log.Logger
	dataPath   string
	pluginPath string

	configGroup   *ServerConfigGroup
	language      *lang.Language
	forceLanguage bool

	memoryManager *MemoryManager
	asyncPool     *scheduler.AsyncPool

	operators *utils.Config
	whitelist *utils.Config
	banByName *permission.BanList
	banByIP   *permission.BanList

	maxPlayers int
	onlineMode bool

	serverID uuid.UUID

	network         *network.Network
	commandMap      *command.SimpleCommandMap
	resourceManager *resourcepacks.ResourcePackManager
	craftingManager *crafting.CraftingManager
	worldManager    *world.WorldManager
	queryInfo       *query.QueryInfo

	playerDataProvider player.PlayerDataProvider

	// playerList is Server::$playerList (raw UUID => Player), in join order.
	playerList []*player.Player

	// broadcastSubscribers is channel ID => subscribers, in subscription order.
	broadcastSubscribers map[string][]command.Sender

	console       *console.ConsoleReader
	consoleSender *console.ConsoleCommandSender

	tickCounter int64
	nextTick    time.Time
	tickAverage [TargetTicksPerSecond]float64
	useAverage  [TargetTicksPerSecond]float64
	currentTPS  float64
	currentUse  float64
	startTime   time.Time

	doTitleTick bool

	// isRunning is Server::$isRunning. It's atomic because Shutdown is called both from inside
	// the tick (the stop command, with the server lock held) and from other goroutines (signals).
	isRunning  atomic.Bool
	hasStopped bool
	stopOnce   sync.Once
	stopped    chan struct{}
}

// knownBlocks are the blocks worlds can hold and send to clients: those with a network serializer
// (network/mcpe/convert/vanilla_block_mappings.go). Keep them in sync until every block is mapped
// (see AGENTS.md §6 Phase 1).
func knownBlocks() []block.Behavior {
	return []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
		block.VanillaGravel(),
		block.VanillaCoalOre(),
		block.VanillaIronOre(),
		block.VanillaRedstoneOre(),
		block.VanillaLapisLazuliOre(),
		block.VanillaGoldOre(),
		block.VanillaDiamondOre(),
		block.VanillaEmeraldOre(),
		block.VanillaWater(),
		block.VanillaSand(),
		block.VanillaSandstone(),
		block.VanillaSnowLayer(),
		block.VanillaTallGrass(),
		block.VanillaOakLog(),
		block.VanillaOakLeaves(),
		block.VanillaSpruceLog(),
		block.VanillaSpruceLeaves(),
		block.VanillaBirchLog(),
		block.VanillaBirchLeaves(),
	}
}

// New is a port of Server::__construct up to startupPrepareNetworkInterfaces: it prepares the data
// folder, loads the configuration, language, lists and the worlds. Start opens the network and
// runs the tick loop (the rest of the constructor).
func New(dataPath string, logger log.Logger) (*Server, error) {
	return NewWithPluginPath(dataPath, filepath.Join(dataPath, "plugins"), logger)
}

// NewWithPluginPath is New with PocketMine.php's --plugins option.
func NewWithPluginPath(dataPath, pluginPath string, logger log.Logger) (*Server, error) {
	s := &Server{
		logger:               logger,
		stopped:              make(chan struct{}),
		startTime:            time.Now(),
		currentTPS:           TargetTicksPerSecond,
		broadcastSubscribers: map[string][]command.Sender{},
	}
	for i := range s.tickAverage {
		s.tickAverage[i] = TargetTicksPerSecond
	}

	timings.Init()

	for _, dir := range []string{dataPath, pluginPath, filepath.Join(dataPath, "worlds"), filepath.Join(dataPath, "players")} {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return nil, fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	s.dataPath, _ = filepath.Abs(dataPath)
	s.pluginPath, _ = filepath.Abs(pluginPath)

	logger.Info("Loading server configuration")
	pocketmineYmlPath := filepath.Join(s.dataPath, "pocketmine.yml")
	if _, err := os.Stat(pocketmineYmlPath); os.IsNotExist(err) {
		content := string(defaultPocketmineYml)
		if pocketmine.IsDevelopmentBuild {
			content = strings.Replace(content, "preferred-channel: stable", "preferred-channel: beta", 1)
		}
		_ = os.WriteFile(pocketmineYmlPath, []byte(content), 0o644)
	}
	pocketmineYml, err := utils.NewConfig(pocketmineYmlPath, utils.ConfigYAML, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("loading pocketmine.yml: %w", err)
	}
	serverProperties, err := utils.NewConfig(filepath.Join(s.dataPath, "server.properties"), utils.ConfigProperties, map[string]any{
		PropertyMotd:                          DefaultServerName,
		PropertyServerPortIPv4:                DefaultPortIPv4,
		PropertyServerPortIPv6:                DefaultPortIPv6,
		PropertyEnableIPv6:                    true,
		PropertyWhitelist:                     false,
		PropertyMaxPlayers:                    DefaultMaxPlayers,
		PropertyGameMode:                      "SURVIVAL", //TODO: this probably shouldn't use the enum name directly
		PropertyForceGameMode:                 false,
		PropertyHardcore:                      false,
		PropertyPvp:                           true,
		PropertyDifficulty:                    world.DifficultyNormal,
		PropertyDefaultWorldGeneratorSettings: "",
		PropertyDefaultWorldName:              "world",
		PropertyDefaultWorldSeed:              "",
		PropertyDefaultWorldGenerator:         "DEFAULT",
		PropertyEnableQuery:                   true,
		PropertyAutoSave:                      true,
		PropertyViewDistance:                  DefaultMaxViewDistance,
		PropertyXboxAuth:                      true,
		PropertyLanguage:                      "eng",
	})
	if err != nil {
		return nil, fmt.Errorf("loading server.properties: %w", err)
	}
	s.configGroup = NewServerConfigGroup(pocketmineYml, serverProperties, defaultArgs())

	debugLogLevel := s.configGroup.GetPropertyInt(YmlDebugLevel, 1)
	if l, ok := logger.(*utils.MainLogger); ok {
		l.SetLogDebug(debugLogLevel > 1)
	}

	s.forceLanguage = s.configGroup.GetPropertyBool(YmlSettingsForceLanguage, false)
	selectedLang := s.configGroup.GetConfigString(PropertyLanguage, s.configGroup.GetPropertyString("settings.language", lang.FallbackLanguage))
	if s.language, err = lang.NewLanguage(selectedLang, "", ""); err != nil {
		logger.Error(err.Error())
		if s.language, err = lang.NewLanguage(lang.FallbackLanguage, "", ""); err != nil {
			return nil, fmt.Errorf("fallback language %q not found", lang.FallbackLanguage)
		}
	}
	logger.Info(s.language.Translate(lang.KnownTranslationFactory.LanguageSelected(s.language.Name(), s.language.Lang())))

	if pocketmine.IsDevelopmentBuild {
		// PHP refuses to start development builds unless settings.enable-dev-builds is set. Every
		// build of this port is a development build (it isn't released), so only the warning is
		// kept.
		logger.Warning(strings.Repeat("-", 40))
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerDevBuildWarning1(pocketmine.Name)))
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerDevBuildWarning2()))
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerDevBuildWarning3()))
		logger.Warning(strings.Repeat("-", 40))
	}

	s.memoryManager = NewMemoryManager(s)

	logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerStart(utils.Aqua + s.GetVersion() + utils.Reset)))

	poolSize := 2
	if v := s.configGroup.GetPropertyString(YmlSettingsAsyncWorkers, "auto"); v == "auto" {
		if processors := runtime.NumCPU() - 2; processors > 0 {
			poolSize = max(1, processors)
		}
	} else {
		n, _ := strconv.Atoi(v)
		poolSize = max(1, n)
	}

	timings.SetEnabled(s.configGroup.GetPropertyBool(YmlSettingsEnableProfiling, false))

	s.asyncPool = scheduler.NewAsyncPool(poolSize, logger)
	workerStartHook := func(i int) {
		if timings.IsEnabled() {
			s.asyncPool.SubmitTaskToWorker(scheduler.NewTimingsControlTaskSetEnabled(true), i)
		}
	}
	s.asyncPool.AddWorkerStartHook(&workerStartHook)

	s.doTitleTick = s.configGroup.GetPropertyBool(YmlConsoleTitleTick, true) && utils.IsTerminalInit() && utils.HasFormattingCodes()

	if s.operators, err = utils.NewConfig(filepath.Join(s.dataPath, "ops.txt"), utils.ConfigEnum, nil); err != nil {
		return nil, err
	}
	if s.whitelist, err = utils.NewConfig(filepath.Join(s.dataPath, "white-list.txt"), utils.ConfigEnum, nil); err != nil {
		return nil, err
	}

	bannedTxt := filepath.Join(s.dataPath, "banned.txt")
	bannedPlayersTxt := filepath.Join(s.dataPath, "banned-players.txt")
	if fileExists(bannedTxt) && !fileExists(bannedPlayersTxt) {
		_ = os.Rename(bannedTxt, bannedPlayersTxt)
	}
	touch(bannedPlayersTxt)
	s.banByName = permission.NewBanList(bannedPlayersTxt)
	s.banByName.Load()
	bannedIpsTxt := filepath.Join(s.dataPath, "banned-ips.txt")
	touch(bannedIpsTxt)
	s.banByIP = permission.NewBanList(bannedIpsTxt)
	s.banByIP.Load()

	s.maxPlayers = s.configGroup.GetConfigInt(PropertyMaxPlayers, DefaultMaxPlayers)

	s.onlineMode = s.configGroup.GetConfigBool(PropertyXboxAuth, true)
	if s.onlineMode {
		logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerAuthEnabled()))
	} else {
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerAuthDisabled()))
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerAuthWarning()))
		logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerAuthPropertyDisabled()))
	}

	if s.configGroup.GetConfigBool(PropertyHardcore, false) && s.GetDifficulty() < world.DifficultyHard {
		s.configGroup.SetConfigInt(PropertyDifficulty, world.DifficultyHard)
	}

	s.serverID = utils.GetMachineUniqueID(s.GetIp() + strconv.Itoa(s.GetPort()))
	logger.Debug("Server unique id: " + s.serverID.String())
	logger.Debug("Machine unique id: " + utils.GetMachineUniqueID("").String())

	s.network = network.NewNetwork(logger)
	s.network.SetName(s.GetMotd())

	versionColor := ""
	if pocketmine.IsDevelopmentBuild {
		versionColor = utils.Yellow
	}
	logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerInfo(s.GetName(), versionColor+s.GetPocketMineVersion()+utils.Reset)))
	logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerLicense(s.GetName())))

	permission.RegisterCorePermissions()

	s.commandMap = command.NewSimpleCommandMap(s)
	RegisterDefaultCommandsFunc(s.commandMap)

	if s.craftingManager, err = crafting.MakeCraftingManager(bedrock.Recipes, bedrock.RecipesDir); err != nil {
		return nil, fmt.Errorf("loading recipes: %w", err)
	}

	if s.resourceManager, err = resourcepacks.NewResourcePackManager(filepath.Join(s.dataPath, "resource_packs"), logger); err != nil {
		return nil, err
	}

	s.worldManager = world.NewWorldManager(filepath.Join(s.dataPath, "worlds"), convert.NewBlockTranslator(), knownBlocks())
	s.worldManager.SetLogger(logger)
	s.worldManager.SetAsyncPool(s.asyncPool)
	s.worldManager.SetPopulationQueueSize(s.configGroup.GetPropertyInt(YmlChunkGenerationPopulationQueueSize, 2))
	s.worldManager.SetAutoSave(s.configGroup.GetConfigBool(PropertyAutoSave, s.worldManager.GetAutoSave()))
	if err := s.worldManager.SetAutoSaveInterval(int64(s.configGroup.GetPropertyInt(YmlTicksPerAutosave, int(s.worldManager.GetAutoSaveInterval())))); err != nil {
		logger.Warning(err.Error())
	}

	s.queryInfo = query.NewQueryInfo(s)

	s.playerDataProvider = player.NewDatFilePlayerDataProvider(filepath.Join(s.dataPath, "players"))

	if !s.startupPrepareWorlds() {
		return nil, fmt.Errorf("%s", s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelDefaultError()))
	}
	// PluginEnableOrder::POSTWORLD: registerServerAliases (no plugins to enable).
	s.commandMap.RegisterServerAliases()
	return s, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func touch(path string) {
	if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		_ = f.Close()
	}
}

// RegisterDefaultCommandsFunc is SimpleCommandMap::setDefaultCommands, installed by the
// command/defaults package (which imports this one's dependencies but not this package).
var RegisterDefaultCommandsFunc = func(m *command.SimpleCommandMap) {}

// getGenerator is Server::startupPrepareWorlds' $getGenerator: nil (after logging why) when the
// generator is unknown or its options are invalid.
func (s *Server) getGenerator(generatorName, generatorOptions, worldName string) generator.Factory {
	factory, ok := generator.GetFactory(strings.ToLower(generatorName))
	if !ok {
		s.logger.Error(s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelGenerationError(
			worldName,
			lang.KnownTranslationFactory.PocketmineLevelUnknownGenerator(generatorName),
		)))
		return nil
	}
	if _, err := factory(0, generatorOptions); err != nil {
		s.logger.Error(s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelGenerationError(
			worldName,
			lang.KnownTranslationFactory.PocketmineLevelInvalidGeneratorOptions(generatorOptions, generatorName, err.Error()),
		)))
		return nil
	}
	return factory
}

// generateWorld is WorldManager::generateWorld with the generator looked up by name, plus the
// spawn terrain pre-generation (see generateSpawnTerrain).
func (s *Server) generateWorld(name string, factory generator.Factory, options *world.WorldCreationOptions) bool {
	gen, err := factory(options.Seed, options.GeneratorOptions)
	if err != nil {
		s.logger.Error(err.Error())
		return false
	}
	w, err := s.worldManager.GenerateWorld(name, gen, options)
	if err != nil {
		s.logger.Error(err.Error())
		return false
	}
	s.generateSpawnTerrain(w)
	return true
}

// startupPrepareWorlds is a port of Server::startupPrepareWorlds.
func (s *Server) startupPrepareWorlds() bool {
	anyWorldFailedToLoad := false

	if worlds, ok := s.configGroup.GetProperty(YmlWorlds, map[string]any{}).(map[string]any); ok {
		names := make([]string, 0, len(worlds))
		for name := range worlds {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, name := range names {
			options, ok := worlds[name].(map[string]any)
			if worlds[name] == nil {
				options = map[string]any{}
			} else if !ok {
				//TODO: this probably should be an error
				continue
			}
			if _, err := s.worldManager.LoadWorld(name); err == nil {
				continue
			}
			if s.worldManager.IsWorldGenerated(name) {
				//allow checking if other worlds are loadable, so the user gets all the errors in one go
				anyWorldFailedToLoad = true
				continue
			}
			creationOptions := world.NewWorldCreationOptions()
			//TODO: error checking

			generatorName := "default"
			if g, ok := options["generator"].(string); ok {
				generatorName = g
			}
			generatorOptions := ""
			if p, ok := options["preset"].(string); ok {
				generatorOptions = p
			}

			factory := s.getGenerator(generatorName, generatorOptions, name)
			if factory == nil {
				anyWorldFailedToLoad = true
				continue
			}
			creationOptions.GeneratorName = strings.ToLower(generatorName)
			creationOptions.GeneratorOptions = generatorOptions

			creationOptions.Difficulty = s.GetDifficulty()
			if d, ok := options["difficulty"].(string); ok {
				creationOptions.Difficulty = world.GetDifficultyFromString(d)
			}

			if seed, ok := options["seed"]; ok {
				if converted, ok := generator.ConvertSeed(fmt.Sprint(seed)); ok {
					creationOptions.Seed = converted
				}
			}

			s.generateWorld(name, factory, creationOptions)
		}
	}

	if s.worldManager.GetDefaultWorld() == nil {
		defaultName := s.configGroup.GetConfigString(PropertyDefaultWorldName, "world")
		if strings.TrimSpace(defaultName) == "" {
			s.logger.Warning("level-name cannot be null, using default")
			defaultName = "world"
			s.configGroup.SetConfigString(PropertyDefaultWorldName, "world")
		}
		if _, err := s.worldManager.LoadWorld(defaultName); err != nil {
			if s.worldManager.IsWorldGenerated(defaultName) {
				s.logger.Error(err.Error())
				s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelDefaultError()))
				return false
			}
			generatorName := s.configGroup.GetConfigString(PropertyDefaultWorldGenerator, "")
			generatorOptions := s.configGroup.GetConfigString(PropertyDefaultWorldGeneratorSettings, "")
			factory := s.getGenerator(generatorName, generatorOptions, defaultName)
			if factory == nil {
				s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelDefaultError()))
				return false
			}
			creationOptions := world.NewWorldCreationOptions()
			creationOptions.GeneratorName = strings.ToLower(generatorName)
			creationOptions.GeneratorOptions = generatorOptions
			if seed, ok := generator.ConvertSeed(s.configGroup.GetConfigString(PropertyDefaultWorldSeed, "")); ok {
				creationOptions.Seed = seed
			}
			creationOptions.Difficulty = s.GetDifficulty()
			if !s.generateWorld(defaultName, factory, creationOptions) {
				s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineLevelDefaultError()))
				return false
			}
		}

		w, ok := s.worldManager.GetWorldByName(defaultName)
		if !ok {
			panic("We just loaded/generated the default world, so it must exist")
		}
		s.worldManager.SetDefaultWorld(w)
	}

	return !anyWorldFailedToLoad
}

// spawnTerrainRadius is the chunk radius WorldManager::generateWorld pre-generates around spawn.
const spawnTerrainRadius = 8

// generateSpawnTerrain is the $backgroundGeneration branch of WorldManager::generateWorld: the
// chunks around the new world's spawn are ordered for population (on the async workers) with
// progress logging. It lives here rather than in WorldManager because player.SelectChunks
// (ChunkSelector) can't be imported from the world package (import cycle). PHP lets the server
// start meanwhile; this port waits for it (collecting the async results itself, since the tick
// isn't running yet), because gophertunnel spawns the client before its chunks can be sent (see
// NetworkSession.OnClientRequestChunkRadius).
func (s *Server) generateSpawnTerrain(w *world.World) {
	s.logger.Notice(fmt.Sprintf("Spawn terrain for world %q is being pregenerated in the background", w.GetFolderName()))

	spawn := w.GetSpawnLocation()
	var selected [][2]int
	for chunk := range player.SelectChunks(spawnTerrainRadius, spawn.FloorX()>>4, spawn.FloorZ()>>4) {
		selected = append(selected, chunk)
	}
	total := len(selected)
	done := 0
	for _, chunk := range selected {
		w.OrderChunkPopulation(chunk[0], chunk[1], nil).OnCompletion(
			func(*format.Chunk) {
				done++
				oldProgress, newProgress := (done-1)*100/total, done*100/total
				if oldProgress/10 != newProgress/10 || done == total || done == 1 {
					s.logger.Info(fmt.Sprintf("[World: %s] Spawn terrain generation progress: %d / %d (%d%%)", w.GetFolderName(), done, total, newProgress))
				}
			},
			func() {
				s.logger.Warning(fmt.Sprintf("[World: %s] Spawn terrain generation failed for chunk %d %d", w.GetFolderName(), chunk[0], chunk[1]))
				done++
			},
		)
	}
	for done < total {
		more, err := s.asyncPool.CollectTasks()
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
		if !more && done < total {
			// Nothing running but requests still queued behind locks: they're drained as results
			// come in, so this only happens if the queue stalled.
			break
		}
		time.Sleep(time.Millisecond)
	}
}

// startupPrepareConnectableNetworkInterfaces is a port of
// Server::startupPrepareConnectableNetworkInterfaces.
func (s *Server) startupPrepareConnectableNetworkInterfaces(ip string, port int, ipV6, useQuery bool, packetBroadcaster mcpe.PacketBroadcaster, entityEventBroadcaster mcpe.EntityEventBroadcaster) bool {
	prettyIP := ip
	if ipV6 {
		prettyIP = "[" + ip + "]"
	}
	rakLibRegistered, err := s.network.RegisterInterface(raklib.NewRakLibInterface(s, ip, port, ipV6, packetBroadcaster, entityEventBroadcaster))
	if err != nil {
		s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerNetworkStartFailed(ip, strconv.Itoa(port), err.Error())))
		return false
	}
	if rakLibRegistered {
		s.logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerNetworkStart(prettyIP, strconv.Itoa(port))))
	}
	if useQuery {
		if !rakLibRegistered {
			//RakLib would normally handle the transport for Query packets
			//if it's not registered we need to make sure Query still works
			if _, err := s.network.RegisterInterface(query.NewDedicatedQueryNetworkInterface(ip, port, ipV6, log.NewPrefixedLogger(s.logger, "Dedicated Query Interface"))); err != nil {
				s.logger.Error(err.Error())
			}
		}
		s.logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerQueryRunning(prettyIP, strconv.Itoa(port))))
	}
	return true
}

// startupPrepareNetworkInterfaces is a port of Server::startupPrepareNetworkInterfaces.
func (s *Server) startupPrepareNetworkInterfaces() bool {
	useQuery := s.configGroup.GetConfigBool(PropertyEnableQuery, true)

	packetBroadcaster := mcpe.NewStandardPacketBroadcaster()
	entityEventBroadcaster := mcpe.NewStandardEntityEventBroadcaster(packetBroadcaster)

	if !s.startupPrepareConnectableNetworkInterfaces(s.GetIp(), s.GetPort(), false, useQuery, packetBroadcaster, entityEventBroadcaster) ||
		(s.configGroup.GetConfigBool(PropertyEnableIPv6, true) &&
			!s.startupPrepareConnectableNetworkInterfaces(s.GetIpV6(), s.GetPortV6(), true, useQuery, packetBroadcaster, entityEventBroadcaster)) {
		return false
	}

	if useQuery {
		s.network.RegisterRawPacketHandler(query.NewQueryHandler(s, s.logger))
	}

	for _, entry := range s.GetIPBans().GetEntries() {
		s.network.BlockAddress(entry.Name(), -1)
	}

	if s.configGroup.GetPropertyBool(YmlNetworkUpnpForwarding, false) {
		internalIP, err := utils.GetInternalIP()
		var iface *upnp.UPnPNetworkInterface
		if err == nil {
			iface, err = upnp.NewUPnPNetworkInterface(s.logger, internalIP, s.GetPort())
		}
		if err == nil {
			_, err = s.network.RegisterInterface(iface)
		}
		if err != nil {
			s.logger.Error(err.Error())
		}
	}
	return true
}

// Start is the rest of Server::__construct: opens the network interfaces, starts the console and
// runs the tick loop until the server stops (then forceShutdown). It returns after shutdown.
func (s *Server) Start() error {
	s.mu.Lock()
	s.isRunning.Store(true)
	if !s.startupPrepareNetworkInterfaces() {
		s.mu.Unlock()
		s.ForceShutdown()
		return fmt.Errorf("failed to start the network interfaces")
	}

	if err := s.configGroup.Save(); err != nil {
		s.logger.Error(err.Error())
	}

	s.logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerDefaultGameMode(s.GetGamemode().GetTranslatableName())))
	highlight, reset := utils.Aqua, utils.Reset
	github := pocketmine.GithubURL
	splash := "\n\n"
	for _, link := range []*lang.Translatable{
		lang.KnownTranslationFactory.PocketmineServerUrlDiscord(highlight + "https://discord.pmmp.io" + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlDocs(highlight + "https://doc.pmmp.io" + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlSourceCode(highlight + github + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlFreePlugins(highlight + "https://poggit.pmmp.io/plugins" + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlDonations(highlight + "https://patreon.com/pocketminemp" + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlTranslations(highlight + "https://translate.pocketmine.net" + reset),
		lang.KnownTranslationFactory.PocketmineServerUrlBugReporting(highlight + github + "/issues" + reset),
	} {
		splash += "- " + s.language.Translate(link) + "\n"
	}
	s.logger.Info(splash)

	s.logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerStartFinished(strconv.FormatFloat(round(time.Since(s.startTime).Seconds(), 3), 'f', -1, 64))))

	forwarder := command.NewBroadcastLoggerForwarder(s, s.logger, s.language)
	s.SubscribeToBroadcastChannel(BroadcastChannelAdministrative, forwarder)
	s.SubscribeToBroadcastChannel(BroadcastChannelUsers, forwarder)

	//TODO: move console parts to a separate component
	if s.configGroup.GetPropertyBool(YmlConsoleEnableInput, true) {
		s.console = console.NewConsoleReader(s.logger)
	}
	s.mu.Unlock()

	s.tickProcessor()
	s.ForceShutdown()
	return nil
}

// ErrorLog is where gophertunnel's and go-raknet's own errors go (see slog_handler.go).
func (s *Server) ErrorLog() *slog.Logger {
	return newSlogLogger(log.NewPrefixedLogger(s.logger, "Network"))
}

// tickProcessor is a port of Server::tickProcessor.
func (s *Server) tickProcessor() {
	s.nextTick = time.Now()
	for {
		s.mu.Lock()
		running := s.isRunning.Load()
		if running {
			s.tick()
		}
		next := s.nextTick
		s.mu.Unlock()
		if !running {
			return
		}
		//sleeps are self-correcting - if we undersleep 1ms on this tick, we'll sleep an extra ms on the next tick
		if d := time.Until(next); d > 0 {
			select {
			case <-time.After(d):
			case <-s.stopped:
			}
		}
	}
}

// tick is a port of Server::tick. The server lock must be held.
func (s *Server) tick() {
	tickTime := time.Now()
	if tickTime.Sub(s.nextTick) < -25*time.Millisecond { //Allow half a tick of diff
		return
	}

	timings.ServerTick.StartTiming()

	s.tickCounter++

	// PluginManager::tickSchedulers: plugins aren't ported, so there are no schedulers to tick.

	timings.SchedulerAsync.StartTiming()
	if _, err := s.asyncPool.CollectTasks(); err != nil {
		s.logger.Error(err.Error())
	}
	timings.SchedulerAsync.StopTiming()

	s.worldManager.Tick(s.tickCounter)

	timings.Connection.StartTiming()
	s.network.Tick()
	timings.Connection.StopTiming()

	if s.tickCounter%TargetTicksPerSecond == 0 {
		if s.doTitleTick {
			s.titleTick()
		}
		s.currentTPS = TargetTicksPerSecond
		s.currentUse = 0

		queryRegenerateEvent := serverevent.NewQueryRegenerateEvent(query.NewQueryInfo(s))
		event.Call(queryRegenerateEvent)
		if info, ok := queryRegenerateEvent.GetQueryInfo().(*query.QueryInfo); ok {
			s.queryInfo = info
		}

		s.network.UpdateName()
		s.network.GetBandwidthTracker().RotateAverageHistory()
	}

	// World::clearCache (every TICKS_PER_WORLD_CACHE_CLEAR): this port's worlds have no block
	// cache to clear.

	if s.tickCounter%ticksPerTPSOverloadWarning == 0 && s.GetTicksPerSecondAverage() < tpsOverloadWarningThreshold {
		s.logger.Warning(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerTickOverload()))
	}

	s.memoryManager.Check()

	if s.console != nil {
		timings.ServerCommand.StartTiming()
		for {
			line, ok := s.console.ReadLine()
			if !ok {
				break
			}
			if s.consoleSender == nil {
				s.consoleSender = console.NewConsoleCommandSender(s, s.language)
			}
			s.DispatchCommand(s.consoleSender, line, false)
		}
		timings.ServerCommand.StopTiming()
	}

	timings.ServerTick.StopTiming()

	totalTickTimeSeconds := time.Since(tickTime).Seconds()
	s.currentTPS = min(TargetTicksPerSecond, 1/max(0.001, totalTickTimeSeconds))
	s.currentUse = min(1, totalTickTimeSeconds/TargetSecondsPerTick)

	timings.Tick(s.currentTPS <= float64(s.configGroup.GetPropertyInt(YmlSettingsProfileReportTrigger, TargetTicksPerSecond)))

	idx := s.tickCounter % TargetTicksPerSecond
	s.tickAverage[idx] = s.currentTPS
	s.useAverage[idx] = s.currentUse

	if s.nextTick.Sub(tickTime) < -time.Second {
		s.nextTick = tickTime
	} else {
		s.nextTick = s.nextTick.Add(time.Duration(TargetSecondsPerTick * float64(time.Second)))
	}
}

// titleTick is a port of Server::titleTick: the console window title.
func (s *Server) titleTick() {
	timings.TitleTick.StartTiming()
	defer timings.TitleTick.StopTiming()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	usage := fmt.Sprintf("%g/%g MB @ %d goroutines", round(float64(m.HeapAlloc)/1024/1024, 2), round(float64(m.Sys)/1024/1024, 2), runtime.NumGoroutine())

	online := len(s.playerList)
	connecting := s.network.GetConnectionCount() - online
	bandwidthStats := s.network.GetBandwidthTracker()

	connectingText := ""
	if connecting > 0 {
		connectingText = fmt.Sprintf(" (+%d connecting)", connecting)
	}
	fmt.Printf("\x1b]0;%s %s | Online %d/%d%s | Memory %s | U %g D %g kB/s | TPS %g | Load %g%%\x07",
		s.GetName(), s.GetPocketMineVersion(), online, s.maxPlayers, connectingText, usage,
		round(bandwidthStats.GetSend().GetAverageBytes()/1024, 2), round(bandwidthStats.GetReceive().GetAverageBytes()/1024, 2),
		s.GetTicksPerSecondAverage(), s.GetTickUsageAverage())
}

// Shutdown is a port of Server::shutdown: the tick loop stops, then forceShutdown runs. It's safe
// to call with or without the server lock held.
func (s *Server) Shutdown() {
	s.isRunning.Store(false)
	s.stopOnce.Do(func() { close(s.stopped) })
}

// ForceShutdown is a port of Server::forceShutdown.
func (s *Server) ForceShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hasStopped {
		return
	}
	if s.doTitleTick {
		fmt.Print("\x1b]0;\x07")
	}
	if s.isRunning.Load() {
		s.logger.Emergency(s.language.Translate(lang.KnownTranslationFactory.PocketmineServerForcingShutdown()))
	}
	s.hasStopped = true
	s.isRunning.Store(false)
	s.stopOnce.Do(func() { close(s.stopped) })

	if s.network != nil {
		s.network.GetSessionManager().Close(s.configGroup.GetPropertyString(YmlSettingsShutdownMessage, "Server closed"), nil)
	}

	if s.worldManager != nil {
		s.logger.Debug("Unloading all worlds")
		for _, w := range s.worldManager.GetWorlds() {
			if _, err := s.worldManager.UnloadWorld(w, true); err != nil {
				s.logger.Error(err.Error())
			}
		}
	}

	s.logger.Debug("Removing event handlers")
	event.Global().UnregisterAll()

	if s.asyncPool != nil {
		s.logger.Debug("Shutting down async task worker pool")
		s.asyncPool.Shutdown()
	}

	if s.configGroup != nil {
		s.logger.Debug("Saving properties")
		if err := s.configGroup.Save(); err != nil {
			s.logger.Error(err.Error())
		}
	}

	if s.console != nil {
		s.logger.Debug("Closing console")
		s.console.Quit()
	}

	if s.network != nil {
		s.logger.Debug("Stopping network interfaces")
		for _, iface := range s.network.GetInterfaces() {
			s.logger.Debug(fmt.Sprintf("Stopping network interface %T", iface))
			if err := s.network.UnregisterInterface(iface); err != nil {
				s.logger.Error(err.Error())
			}
		}
	}
}

// Lock/Unlock implement mcpe.Server's Locker.
func (s *Server) Lock()   { s.mu.Lock() }
func (s *Server) Unlock() { s.mu.Unlock() }

// GetName is a port of Server::getName.
func (s *Server) GetName() string { return pocketmine.Name }

// IsRunning is a port of Server::isRunning.
func (s *Server) IsRunning() bool { return s.isRunning.Load() }

// GetPocketMineVersion is a port of Server::getPocketMineVersion.
func (s *Server) GetPocketMineVersion() string { return pocketmine.Version().GetFullVersion(true) }

// GetVersion is a port of Server::getVersion: the Minecraft version.
func (s *Server) GetVersion() string { return protocol.CurrentVersion }

// GetApiVersion is a port of Server::getApiVersion.
func (s *Server) GetApiVersion() string { return pocketmine.BaseVersion }

func (s *Server) GetDataPath() string   { return s.dataPath }
func (s *Server) GetPluginPath() string { return s.pluginPath }

// GetMaxPlayers is a port of Server::getMaxPlayers.
func (s *Server) GetMaxPlayers() int { return s.maxPlayers }

// GetOnlineMode is a port of Server::getOnlineMode.
func (s *Server) GetOnlineMode() bool { return s.onlineMode }

// RequiresAuthentication is a port of Server::requiresAuthentication.
func (s *Server) RequiresAuthentication() bool { return s.GetOnlineMode() }

// GetPort is a port of Server::getPort.
func (s *Server) GetPort() int {
	return s.configGroup.GetConfigInt(PropertyServerPortIPv4, DefaultPortIPv4)
}

// GetPortV6 is a port of Server::getPortV6.
func (s *Server) GetPortV6() int {
	return s.configGroup.GetConfigInt(PropertyServerPortIPv6, DefaultPortIPv6)
}

// GetViewDistance is a port of Server::getViewDistance.
func (s *Server) GetViewDistance() int {
	return max(2, s.configGroup.GetConfigInt(PropertyViewDistance, DefaultMaxViewDistance))
}

// GetAllowedViewDistance is a port of Server::getAllowedViewDistance.
func (s *Server) GetAllowedViewDistance(distance int) int {
	return max(2, min(distance, s.memoryManager.GetViewDistance(s.GetViewDistance())))
}

// GetIp is a port of Server::getIp.
func (s *Server) GetIp() string {
	if str := s.configGroup.GetConfigString(PropertyServerIPv4, ""); str != "" {
		return str
	}
	return "0.0.0.0"
}

// GetIpV6 is a port of Server::getIpV6.
func (s *Server) GetIpV6() string {
	if str := s.configGroup.GetConfigString(PropertyServerIPv6, ""); str != "" {
		return str
	}
	return "::"
}

// GetServerUniqueId is a port of Server::getServerUniqueId.
func (s *Server) GetServerUniqueId() uuid.UUID { return s.serverID }

// GetGamemode is a port of Server::getGamemode.
func (s *Server) GetGamemode() player.GameMode {
	if g, ok := player.GameModeFromString(s.configGroup.GetConfigString(PropertyGameMode, "")); ok {
		return g
	}
	return player.GameModeSurvival
}

// GetForceGamemode is a port of Server::getForceGamemode.
func (s *Server) GetForceGamemode() bool {
	return s.configGroup.GetConfigBool(PropertyForceGameMode, false)
}

// GetDifficulty is a port of Server::getDifficulty.
func (s *Server) GetDifficulty() int {
	return s.configGroup.GetConfigInt(PropertyDifficulty, world.DifficultyNormal)
}

// HasWhitelist is a port of Server::hasWhitelist.
func (s *Server) HasWhitelist() bool { return s.configGroup.GetConfigBool(PropertyWhitelist, false) }

// IsHardcore is a port of Server::isHardcore.
func (s *Server) IsHardcore() bool { return s.configGroup.GetConfigBool(PropertyHardcore, false) }

// GetMotd is a port of Server::getMotd.
func (s *Server) GetMotd() string {
	return s.configGroup.GetConfigString(PropertyMotd, DefaultServerName)
}

func (s *Server) GetLogger() log.Logger { return s.logger }

// GetCraftingManager is a port of Server::getCraftingManager.
func (s *Server) GetCraftingManager() *crafting.CraftingManager { return s.craftingManager }

func (s *Server) GetResourcePackManager() *resourcepacks.ResourcePackManager {
	return s.resourceManager
}
func (s *Server) GetWorldManager() *world.WorldManager           { return s.worldManager }
func (s *Server) GetAsyncPool() *scheduler.AsyncPool             { return s.asyncPool }
func (s *Server) GetTick() int64                                 { return s.tickCounter }
func (s *Server) GetStartTime() time.Time                        { return s.startTime }
func (s *Server) GetCommandMap() command.CommandMap              { return s.commandMap }
func (s *Server) GetSimpleCommandMap() *command.SimpleCommandMap { return s.commandMap }
func (s *Server) GetConfigGroup() *ServerConfigGroup             { return s.configGroup }
func (s *Server) GetNameBans() *permission.BanList               { return s.banByName }
func (s *Server) GetIPBans() *permission.BanList                 { return s.banByIP }
func (s *Server) GetWhitelisted() *utils.Config                  { return s.whitelist }
func (s *Server) GetOps() *utils.Config                          { return s.operators }
func (s *Server) GetLanguage() *lang.Language                    { return s.language }
func (s *Server) IsLanguageForced() bool                         { return s.forceLanguage }
func (s *Server) GetNetwork() *network.Network                   { return s.network }
func (s *Server) GetMemoryManager() *MemoryManager               { return s.memoryManager }
func (s *Server) GetQueryInformation() *query.QueryInfo          { return s.queryInfo }
func (s *Server) GetPropertyInt(variable string, defaultValue int) int {
	return s.configGroup.GetPropertyInt(variable, defaultValue)
}
func (s *Server) GetPropertyBool(variable string, defaultValue bool) bool {
	return s.configGroup.GetPropertyBool(variable, defaultValue)
}
func (s *Server) GetConfigBool(variable string, defaultValue bool) bool {
	return s.configGroup.GetConfigBool(variable, defaultValue)
}

// GetTicksPerSecond is a port of Server::getTicksPerSecond.
func (s *Server) GetTicksPerSecond() float64 { return round(s.currentTPS, 2) }

// GetTicksPerSecondAverage is a port of Server::getTicksPerSecondAverage.
func (s *Server) GetTicksPerSecondAverage() float64 {
	sum := 0.0
	for _, v := range s.tickAverage {
		sum += v
	}
	return round(sum/float64(len(s.tickAverage)), 2)
}

// GetTickUsage is a port of Server::getTickUsage.
func (s *Server) GetTickUsage() float64 { return round(s.currentUse*100, 2) }

// GetTickUsageAverage is a port of Server::getTickUsageAverage.
func (s *Server) GetTickUsageAverage() float64 {
	sum := 0.0
	for _, v := range s.useAverage {
		sum += v
	}
	return round(sum/float64(len(s.useAverage))*100, 2)
}

// round is PHP's round($x, $precision).
func round(x float64, precision int) float64 {
	p := stdmath.Pow(10, float64(precision))
	return stdmath.Round(x*p) / p
}

// GetOnlinePlayers is a port of Server::getOnlinePlayers.
func (s *Server) GetOnlinePlayers() []*player.Player {
	return append([]*player.Player(nil), s.playerList...)
}

// ShouldSavePlayerData is a port of Server::shouldSavePlayerData.
func (s *Server) ShouldSavePlayerData() bool {
	return s.configGroup.GetPropertyBool(YmlPlayerSavePlayerData, true)
}

// GetOfflinePlayer is a port of Server::getOfflinePlayer: an online *player.Player or an
// *player.OfflinePlayer.
func (s *Server) GetOfflinePlayer(name string) any {
	name = strings.ToLower(name)
	if p := s.GetPlayerExact(name); p != nil {
		return p
	}
	return player.NewOfflinePlayer(name, s.GetOfflinePlayerData(name))
}

// HasOfflinePlayerData is a port of Server::hasOfflinePlayerData.
func (s *Server) HasOfflinePlayerData(name string) bool { return s.playerDataProvider.HasData(name) }

// GetOfflinePlayerData is a port of Server::getOfflinePlayerData.
func (s *Server) GetOfflinePlayerData(name string) *nbt.CompoundTag {
	timings.SyncPlayerDataLoad.StartTiming()
	defer timings.SyncPlayerDataLoad.StopTiming()
	if !s.playerDataProvider.HasData(name) {
		return nil
	}
	data, err := s.playerDataProvider.LoadData(name)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to load player data for %s: %v", name, err))
		s.logger.Error(s.language.Translate(lang.KnownTranslationFactory.PocketmineDataPlayerCorrupted(name)))
		return nil
	}
	return data
}

// SaveOfflinePlayerData is a port of Server::saveOfflinePlayerData.
func (s *Server) SaveOfflinePlayerData(name string, data *nbt.CompoundTag) {
	var online playerevent.Player
	if p := s.GetPlayerExact(name); p != nil {
		online = p
	}
	ev := playerevent.NewPlayerDataSaveEvent(data, name, online)
	if !s.ShouldSavePlayerData() {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}
	timings.SyncPlayerDataSave.StartTiming()
	defer timings.SyncPlayerDataSave.StopTiming()
	if err := s.playerDataProvider.SaveData(name, ev.GetSaveData()); err != nil {
		s.logger.Critical(s.language.Translate(lang.KnownTranslationFactory.PocketmineDataSaveError(name, err.Error())))
	}
}

// CreatePlayer is a port of Server::createPlayer. World::requestSafeSpawn resolves synchronously
// in this port (generation is synchronous).
func (s *Server) CreatePlayer(session *mcpe.NetworkSession, playerInfo player.Info, authenticated bool, offlinePlayerData *nbt.CompoundTag) (*player.Player, error) {
	ev := playerevent.NewPlayerCreationEvent(session)
	event.Call(ev)

	var location entity.Location
	hasLocation := false
	if offlinePlayerData != nil {
		if w, ok := s.worldManager.GetWorldByName(string(offlinePlayerData.GetStringOr(player.TagLevel, ""))); ok {
			if loc, err := entity.ParseLocation(offlinePlayerData, w); err == nil {
				location, hasLocation = loc, true
			}
		}
	}
	if !hasLocation { //new player or no valid position due to world not being loaded
		w := s.worldManager.GetDefaultWorld()
		if w == nil {
			panic("Default world should always be loaded")
		}
		spawn := w.GetSafeSpawn(w.GetSpawnLocation())
		if !session.IsConnected() {
			return nil, fmt.Errorf("session disconnected while finding a safe spawn")
		}
		location = entity.LocationFromObject(spawn, w, 0, 0)
	}

	p := player.NewPlayer(s, session, playerInfo, authenticated, location, offlinePlayerData)
	if offlinePlayerData == nil {
		p.OnGround = true //TODO: this hack is needed for new players in-air ticks - they don't get detected as on-ground until they move
	}
	return p, nil
}

// GetPlayerByPrefix is a port of Server::getPlayerByPrefix.
func (s *Server) GetPlayerByPrefix(name string) *player.Player {
	var found *player.Player
	name = strings.ToLower(name)
	delta := stdmath.MaxInt
	for _, p := range s.playerList {
		if strings.HasPrefix(strings.ToLower(p.GetName()), name) {
			curDelta := len(p.GetName()) - len(name)
			if curDelta < delta {
				found = p
				delta = curDelta
			}
			if curDelta == 0 {
				break
			}
		}
	}
	return found
}

// GetPlayerExact is a port of Server::getPlayerExact.
func (s *Server) GetPlayerExact(name string) *player.Player {
	name = strings.ToLower(name)
	for _, p := range s.playerList {
		if strings.ToLower(p.GetName()) == name {
			return p
		}
	}
	return nil
}

// GetPlayerByUUID is a port of Server::getPlayerByUUID.
func (s *Server) GetPlayerByUUID(id uuid.UUID) *player.Player {
	for _, p := range s.playerList {
		if p.GetUniqueID() == id {
			return p
		}
	}
	return nil
}

// AddOp is a port of Server::addOp.
func (s *Server) AddOp(name string) {
	s.operators.Set(strings.ToLower(name), true)
	if p := s.GetPlayerExact(name); p != nil {
		p.SetBasePermission(permission.RootOperator, true)
	}
	if err := s.operators.Save(); err != nil {
		s.logger.Error(err.Error())
	}
}

// RemoveOp is a port of Server::removeOp.
func (s *Server) RemoveOp(name string) {
	lowercaseName := strings.ToLower(name)
	for operatorName := range s.operators.GetAll() {
		if lowercaseName == strings.ToLower(operatorName) {
			s.operators.Remove(operatorName)
		}
	}
	if p := s.GetPlayerExact(name); p != nil {
		p.UnsetBasePermission(permission.RootOperator)
	}
	if err := s.operators.Save(); err != nil {
		s.logger.Error(err.Error())
	}
}

// AddWhitelist is a port of Server::addWhitelist.
func (s *Server) AddWhitelist(name string) {
	s.whitelist.Set(strings.ToLower(name), true)
	if err := s.whitelist.Save(); err != nil {
		s.logger.Error(err.Error())
	}
}

// RemoveWhitelist is a port of Server::removeWhitelist.
func (s *Server) RemoveWhitelist(name string) {
	s.whitelist.Remove(strings.ToLower(name))
	if err := s.whitelist.Save(); err != nil {
		s.logger.Error(err.Error())
	}
}

// IsWhitelisted is a port of Server::isWhitelisted.
func (s *Server) IsWhitelisted(name string) bool {
	return !s.HasWhitelist() || s.operators.Exists(name, true) || s.whitelist.Exists(name, true)
}

// IsOp is a port of Server::isOp.
func (s *Server) IsOp(name string) bool { return s.operators.Exists(name, true) }

// GetCommandAliases is a port of Server::getCommandAliases.
func (s *Server) GetCommandAliases() map[string][]string {
	result := map[string][]string{}
	section, ok := s.configGroup.GetProperty(YmlAliases, nil).(map[string]any)
	if !ok {
		return result
	}
	for key, value := range section {
		//TODO: more validation needed here
		//key might not be a string, value might not be list<string>
		var commands []string
		if list, ok := value.([]any); ok {
			for _, c := range list {
				commands = append(commands, fmt.Sprint(c))
			}
		} else {
			commands = append(commands, fmt.Sprint(value))
		}
		result[key] = commands
	}
	return result
}

// SubscribeToBroadcastChannel is a port of Server::subscribeToBroadcastChannel.
func (s *Server) SubscribeToBroadcastChannel(channelID string, subscriber command.Sender) {
	if !slices.Contains(s.broadcastSubscribers[channelID], subscriber) {
		s.broadcastSubscribers[channelID] = append(s.broadcastSubscribers[channelID], subscriber)
	}
}

// UnsubscribeFromBroadcastChannel is a port of Server::unsubscribeFromBroadcastChannel.
func (s *Server) UnsubscribeFromBroadcastChannel(channelID string, subscriber command.Sender) {
	subscribers := slices.DeleteFunc(s.broadcastSubscribers[channelID], func(other command.Sender) bool { return other == subscriber })
	if len(subscribers) == 0 {
		delete(s.broadcastSubscribers, channelID)
	} else {
		s.broadcastSubscribers[channelID] = subscribers
	}
}

// UnsubscribeFromAllBroadcastChannels is a port of Server::unsubscribeFromAllBroadcastChannels.
func (s *Server) UnsubscribeFromAllBroadcastChannels(subscriber command.Sender) {
	for channelID := range s.broadcastSubscribers {
		s.UnsubscribeFromBroadcastChannel(channelID, subscriber)
	}
}

// GetBroadcastChannelSubscribers is a port of Server::getBroadcastChannelSubscribers.
func (s *Server) GetBroadcastChannelSubscribers(channelID string) []command.Sender {
	return append([]command.Sender(nil), s.broadcastSubscribers[channelID]...)
}

// BroadcastMessage is a port of Server::broadcastMessage: nil recipients means the
// BROADCAST_CHANNEL_USERS subscribers.
func (s *Server) BroadcastMessage(message any, recipients []command.Sender) int {
	if recipients == nil {
		recipients = s.GetBroadcastChannelSubscribers(BroadcastChannelUsers)
	}
	for _, recipient := range recipients {
		recipient.SendMessage(message)
	}
	return len(recipients)
}

// getPlayerBroadcastSubscribers is a port of Server::getPlayerBroadcastSubscribers.
func (s *Server) getPlayerBroadcastSubscribers(channelID string) []*player.Player {
	var players []*player.Player
	for _, subscriber := range s.broadcastSubscribers[channelID] {
		if p, ok := subscriber.(*player.Player); ok {
			players = append(players, p)
		}
	}
	return players
}

// BroadcastTip is a port of Server::broadcastTip.
func (s *Server) BroadcastTip(tip string, recipients []*player.Player) int {
	if recipients == nil {
		recipients = s.getPlayerBroadcastSubscribers(BroadcastChannelUsers)
	}
	for _, recipient := range recipients {
		recipient.SendTip(tip)
	}
	return len(recipients)
}

// BroadcastPopup is a port of Server::broadcastPopup.
func (s *Server) BroadcastPopup(popup string, recipients []*player.Player) int {
	if recipients == nil {
		recipients = s.getPlayerBroadcastSubscribers(BroadcastChannelUsers)
	}
	for _, recipient := range recipients {
		recipient.SendPopup(popup)
	}
	return len(recipients)
}

// BroadcastTitle is a port of Server::broadcastTitle.
func (s *Server) BroadcastTitle(title, subtitle string, fadeIn, stay, fadeOut int, recipients []*player.Player) int {
	if recipients == nil {
		recipients = s.getPlayerBroadcastSubscribers(BroadcastChannelUsers)
	}
	for _, recipient := range recipients {
		recipient.SendTitle(title, subtitle, fadeIn, stay, fadeOut)
	}
	return len(recipients)
}

// DispatchCommand is a port of Server::dispatchCommand.
func (s *Server) DispatchCommand(sender command.Sender, commandLine string, internal bool) bool {
	if !internal {
		ev := serverevent.NewCommandEvent(sender, commandLine)
		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
		commandLine = ev.GetCommand()
	}
	return s.commandMap.Dispatch(sender, commandLine)
}

// AddOnlinePlayer is a port of Server::addOnlinePlayer.
func (s *Server) AddOnlinePlayer(p *player.Player) bool {
	ev := playerevent.NewPlayerLoginEvent(p, "Plugin reason")
	event.Call(ev)
	if ev.IsCancelled() || !p.IsConnected() {
		p.Disconnect(ev.GetKickMessage(), nil, nil)
		return false
	}

	session := p.GetNetworkSession()
	pos := p.GetPosition()
	s.logger.Info(s.language.Translate(lang.KnownTranslationFactory.PocketminePlayerLogIn(
		utils.Aqua+p.GetName()+utils.Reset,
		session.GetIp(),
		strconv.Itoa(session.GetPort()),
		strconv.Itoa(p.GetID()),
		p.GetWorld().GetDisplayName(),
		strconv.FormatFloat(round(pos.X, 4), 'f', -1, 64),
		strconv.FormatFloat(round(pos.Y, 4), 'f', -1, 64),
		strconv.FormatFloat(round(pos.Z, 4), 'f', -1, 64),
	)))

	for _, other := range s.playerList {
		if otherSession, ok := other.GetNetworkSession().(*mcpe.NetworkSession); ok {
			otherSession.OnPlayerAdded(p)
		}
	}
	s.playerList = append(s.playerList, p)
	return true
}

// RemoveOnlinePlayer is a port of Server::removeOnlinePlayer.
func (s *Server) RemoveOnlinePlayer(p *player.Player) {
	i := slices.Index(s.playerList, p)
	if i == -1 {
		return
	}
	s.playerList = slices.Delete(s.playerList, i, i+1)
	for _, remaining := range s.playerList {
		if session, ok := remaining.GetNetworkSession().(*mcpe.NetworkSession); ok {
			session.OnPlayerRemoved(p)
		}
	}
}

// query.Server implementation.

func (s *Server) QueryListPlugins() bool {
	return s.configGroup.GetPropertyBool(YmlSettingsQueryPlugins, true)
}

// GetQueryPlugins is PluginManager::getPlugins for QueryInfo: plugins aren't ported.
func (s *Server) GetQueryPlugins() []query.Plugin { return nil }

func (s *Server) GetOnlinePlayerNames() []string {
	names := make([]string, 0, len(s.playerList))
	for _, p := range s.playerList {
		if p.IsOnline() {
			names = append(names, p.GetName())
		}
	}
	return names
}

func (s *Server) IsSurvivalLike() bool {
	g := s.GetGamemode()
	return g == player.GameModeSurvival || g == player.GameModeAdventure
}

func (s *Server) GetDefaultWorldName() (string, bool) {
	if w := s.worldManager.GetDefaultWorld(); w != nil {
		return w.GetDisplayName(), true
	}
	return "", false
}

var (
	_ player.Server  = (*Server)(nil)
	_ raklib.Server  = (*Server)(nil)
	_ query.Server   = (*Server)(nil)
	_ command.Server = (*Server)(nil)
)
