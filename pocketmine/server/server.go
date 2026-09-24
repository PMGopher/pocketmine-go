package server

import (
	"bufio"
	"fmt"
	stdmath "math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	worldio "pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/generator"

	// Linked for their init(): EntityFactory registrations and the world/block hooks that create
	// item entities, experience orbs, falling blocks and primed TNT.
	_ "pocketmine-go/pocketmine/entity/object"
	_ "pocketmine-go/pocketmine/entity/projectile"
	// Linked for its init(), which gives NetworkSession its packet handlers.
	_ "pocketmine-go/pocketmine/network/mcpe/handler"
)

// Server constants, Server::DEFAULT_* and TARGET_*.
const (
	DefaultServerName      = pocketmine.Name + " Server"
	DefaultMaxPlayers      = 20
	DefaultPortIPv4        = 19132
	DefaultPortIPv6        = 19133
	DefaultMaxViewDistance = 16
	TargetTicksPerSecond   = 20
	TargetSecondsPerTick   = 1.0 / TargetTicksPerSecond

	// BroadcastChannelUsers is Server::BROADCAST_CHANNEL_USERS.
	BroadcastChannelUsers = "pocketmine.broadcast.user"
)

// Server is a port of pocketmine\Server: it loads the configuration and the default world, accepts
// connections, runs the tick loop and keeps the online player list.
//
// Not ported yet: plugins, the command map and default commands (the console only knows "stop"),
// the async pool, the Query/UPnP interfaces, crash dumps, timings, the memory manager, ban/op/
// whitelist lists, resource packs, the Language (messages sent to clients stay untranslated keys
// the client translates), and pocketmine.yml (see ServerConfigGroup).
type Server struct {
	// mu serialises the tick with packet handling, which PHP does on one thread (see
	// mcpe.Server's Locker).
	mu sync.Mutex

	logger      log.Logger
	dataPath    string
	configGroup *ServerConfigGroup

	worldManager       *world.WorldManager
	playerDataProvider player.PlayerDataProvider

	listener *minecraft.Listener

	// playerList is Server::$playerList, in join order.
	playerList []*player.Player

	tickCounter int64
	isRunning   bool
	stopOnce    sync.Once
	stopped     chan struct{}
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

// New is a port of Server::__construct: it prepares the data folder, loads server.properties and
// the default world. Call Start to open the network and run the server.
func New(dataPath string, logger log.Logger) (*Server, error) {
	s := &Server{logger: logger, dataPath: dataPath, stopped: make(chan struct{})}

	for _, dir := range []string{dataPath, filepath.Join(dataPath, "worlds"), filepath.Join(dataPath, "players")} {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return nil, fmt.Errorf("creating %s: %w", dir, err)
		}
	}

	logger.Info("Loading server configuration")
	serverProperties, err := utils.NewConfig(filepath.Join(dataPath, "server.properties"), utils.ConfigProperties, map[string]any{
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
		PropertyDifficulty:                    worldio.DifficultyNormal,
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
	s.configGroup = NewServerConfigGroup(nil, serverProperties, defaultArgs())

	logger.Info(fmt.Sprintf("Starting Minecraft: Bedrock Edition server version %s", protocol.CurrentVersion))

	s.playerDataProvider = player.NewDatFilePlayerDataProvider(filepath.Join(dataPath, "players"))

	s.worldManager = world.NewWorldManager(filepath.Join(dataPath, "worlds"), convert.NewBlockTranslator(), knownBlocks())
	s.worldManager.SetAutoSave(s.configGroup.GetConfigBool(PropertyAutoSave, s.worldManager.GetAutoSave()))

	if err := s.startupPrepareWorlds(); err != nil {
		return nil, err
	}
	return s, nil
}

// startupPrepareWorlds is a port of Server::startupPrepareWorlds for the default world (the
// pocketmine.yml "worlds" list isn't ported).
func (s *Server) startupPrepareWorlds() error {
	name := s.configGroup.GetConfigString(PropertyDefaultWorldName, "world")
	if strings.TrimSpace(name) == "" {
		s.logger.Warning("level-name cannot be null, using default")
		name = "world"
		s.configGroup.SetConfigString(PropertyDefaultWorldName, "world")
	}

	w, err := s.worldManager.LoadWorld(name)
	if err != nil {
		if s.worldManager.IsWorldGenerated(name) {
			return fmt.Errorf("could not load the default world %q: %w", name, err)
		}

		generatorName := s.configGroup.GetConfigString(PropertyDefaultWorldGenerator, "")
		factory, ok := generator.GetFactory(strings.ToLower(generatorName))
		if !ok {
			return fmt.Errorf("could not generate world %q: unknown generator %q", name, generatorName)
		}
		options := world.NewWorldCreationOptions()
		options.GeneratorName = strings.ToLower(generatorName)
		options.GeneratorOptions = s.configGroup.GetConfigString(PropertyDefaultWorldGeneratorSettings, "")
		if seed, ok := generator.ConvertSeed(s.configGroup.GetConfigString(PropertyDefaultWorldSeed, "")); ok {
			options.Seed = seed
		}
		options.Difficulty = s.GetDifficulty()

		gen, err := factory(options.Seed, options.GeneratorOptions)
		if err != nil {
			return fmt.Errorf("could not generate world %q: %w", name, err)
		}
		s.logger.Info(fmt.Sprintf("Generating world %q (seed %d)", name, options.Seed))
		if w, err = s.worldManager.GenerateWorld(name, gen, options); err != nil {
			return fmt.Errorf("could not generate world %q: %w", name, err)
		}
	}
	s.worldManager.SetDefaultWorld(w)
	return nil
}

// Start opens the network interface (Server::startupPrepareConnectableNetworkInterfaces with
// gophertunnel in place of RakLib) and runs the server until it's stopped: the tick loop, the
// console reader and the connection accept loop. It returns after Shutdown.
func (s *Server) Start() error {
	port := s.GetPort()
	cfg := minecraft.ListenConfig{
		StatusProvider:         minecraft.NewStatusProvider(s.GetMotd(), pocketmine.Name),
		AuthenticationDisabled: !s.RequiresAuthentication(),
		MaximumPlayers:         s.GetMaxPlayers(),
		ErrorLog:               newSlogLogger(log.NewPrefixedLogger(s.logger, "Network")),
	}
	listener, err := cfg.Listen("raknet", s.GetIp()+":"+strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("failed to bind to port %d: %w", port, err)
	}
	s.listener = listener
	s.isRunning = true

	s.logger.Info(fmt.Sprintf("Minecraft network interface running on %s", listener.Addr()))
	if !s.RequiresAuthentication() {
		s.logger.Warning("Online mode is disabled. The server will not verify that players are authenticated to their Xbox Live accounts.")
	}
	s.logger.Info(fmt.Sprintf("Default game type: %s", s.GetGamemode().GetEnglishName()))
	s.logger.Info(fmt.Sprintf("Done! For help, type \"help\" or \"?\""))

	go s.acceptLoop()
	go s.readConsole()
	s.tickProcessor()
	return nil
}

// acceptLoop hands every accepted connection to its own NetworkSession (NetworkInterface /
// RakLibInterface::onClientConnect).
func (s *Server) acceptLoop() {
	for {
		c, err := s.listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*minecraft.Conn)
		s.logger.Info(fmt.Sprintf("%s connecting from %s", conn.IdentityData().DisplayName, conn.RemoteAddr()))
		go mcpe.NewNetworkSession(s, conn).Run()
	}
}

// readConsole is a stand-in for ConsoleReaderChildProcessDaemon + dispatchCommand: the command
// map and default commands aren't ported, so only "stop" (StopCommand) is understood.
func (s *Server) readConsole() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch strings.ToLower(strings.Fields(line)[0]) {
		case "stop":
			s.logger.Info("Stopping the server...")
			s.Shutdown()
			return
		default:
			s.logger.Info("Unknown command. Try /help for a list of commands")
		}
	}
}

// tickProcessor is a port of Server::tickProcessor: ticks at TargetTicksPerSecond until shutdown.
func (s *Server) tickProcessor() {
	ticker := time.NewTicker(time.Duration(TargetSecondsPerTick * float64(time.Second)))
	defer ticker.Stop()
	for {
		select {
		case <-s.stopped:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

// tick is a port of Server::tick: worlds, then network sessions (chunk sending).
func (s *Server) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isRunning {
		return
	}
	s.tickCounter++
	s.worldManager.Tick(s.tickCounter)
	for _, p := range s.playerList {
		if session, ok := p.GetNetworkSession().(*mcpe.NetworkSession); ok {
			session.Tick()
		}
	}
}

// Shutdown is a port of Server::shutdown + forceShutdown: players are disconnected and saved, then
// worlds are unloaded (saving them) and the configuration is saved.
func (s *Server) Shutdown() {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.isRunning = false

		for _, p := range append([]*player.Player(nil), s.playerList...) {
			if session, ok := p.GetNetworkSession().(*mcpe.NetworkSession); ok {
				session.Disconnect("Server closed")
			}
		}

		s.logger.Info("Stopping network interfaces")
		if s.listener != nil {
			_ = s.listener.Close()
		}

		s.logger.Info("Unloading worlds")
		for _, w := range s.worldManager.GetWorlds() {
			if _, err := s.worldManager.UnloadWorld(w, true); err != nil {
				s.logger.Error(err.Error())
			}
		}

		s.logger.Info("Saving properties")
		if err := s.configGroup.Save(); err != nil {
			s.logger.Error(err.Error())
		}
		close(s.stopped)
	})
}

// Lock/Unlock implement mcpe.Server's Locker.
func (s *Server) Lock()   { s.mu.Lock() }
func (s *Server) Unlock() { s.mu.Unlock() }

func (s *Server) GetLogger() log.Logger                { return s.logger }
func (s *Server) GetDataPath() string                  { return s.dataPath }
func (s *Server) GetConfigGroup() *ServerConfigGroup   { return s.configGroup }
func (s *Server) GetWorldManager() *world.WorldManager { return s.worldManager }
func (s *Server) GetTick() int64                       { return s.tickCounter }
func (s *Server) IsRunning() bool                      { return s.isRunning }
func (s *Server) GetMotd() string {
	return s.configGroup.GetConfigString(PropertyMotd, DefaultServerName)
}
func (s *Server) GetMaxPlayers() int {
	return s.configGroup.GetConfigInt(PropertyMaxPlayers, DefaultMaxPlayers)
}
func (s *Server) GetPort() int {
	return s.configGroup.GetConfigInt(PropertyServerPortIPv4, DefaultPortIPv4)
}
func (s *Server) GetIp() string { return s.configGroup.GetConfigString(PropertyServerIPv4, "") }
func (s *Server) RequiresAuthentication() bool {
	return s.configGroup.GetConfigBool(PropertyXboxAuth, true)
}
func (s *Server) GetForceGamemode() bool {
	return s.configGroup.GetConfigBool(PropertyForceGameMode, false)
}
func (s *Server) GetViewDistance() int {
	return max(2, s.configGroup.GetConfigInt(PropertyViewDistance, DefaultMaxViewDistance))
}

// GetDifficulty is a port of Server::getDifficulty.
func (s *Server) GetDifficulty() int {
	return s.configGroup.GetConfigInt(PropertyDifficulty, worldio.DifficultyNormal)
}

// GetGamemode is a port of Server::getGamemode.
func (s *Server) GetGamemode() player.GameMode {
	if g, ok := player.GameModeFromString(s.configGroup.GetConfigString(PropertyGameMode, "SURVIVAL")); ok {
		return g
	}
	return player.GameModeSurvival
}

// GetDefaultGameMode is the server game mode as the network game mode ID (StartGame's world game
// mode, TypeConverter::coreGameModeToProtocol($server->getGamemode())).
func (s *Server) GetDefaultGameMode() int32 {
	return convert.CoreGameModeToProtocol(int(s.GetGamemode()))
}

// GetAllowedViewDistance is a port of Server::getAllowedViewDistance (the memory manager isn't
// ported, so the configured view distance is the limit).
func (s *Server) GetAllowedViewDistance(distance int) int {
	return max(2, min(distance, s.GetViewDistance()))
}

// GetOfflinePlayerData is a port of Server::getOfflinePlayerData.
func (s *Server) GetOfflinePlayerData(name string) *nbt.CompoundTag {
	name = strings.ToLower(name)
	if !s.playerDataProvider.HasData(name) {
		return nil
	}
	data, err := s.playerDataProvider.LoadData(name)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to load player data for %s: %v", name, err))
		return nil
	}
	return data
}

// SaveOfflinePlayerData is a port of Server::saveOfflinePlayerData (without the cancellable
// PlayerDataSaveEvent).
func (s *Server) SaveOfflinePlayerData(name string, data *nbt.CompoundTag) {
	if err := s.playerDataProvider.SaveData(strings.ToLower(name), data); err != nil {
		s.logger.Critical(fmt.Sprintf("Could not save player data for %s: %v", name, err))
	}
}

// CreatePlayer is a port of Server::createPlayer: returning players continue where they were, new
// ones start at the default world's safe spawn (World::requestSafeSpawn).
func (s *Server) CreatePlayer(session *mcpe.NetworkSession, info *player.XboxLivePlayerInfo) (*player.Player, error) {
	offlinePlayerData := s.GetOfflinePlayerData(info.GetUsername())

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
			return nil, fmt.Errorf("default world should always be loaded")
		}
		spawn := w.GetSafeSpawn(w.GetSpawnLocation())
		location = entity.LocationFromObject(spawn, w, 0, 0)
	}

	playerUUID, err := uuid.Parse(info.GetUUID())
	if err != nil {
		return nil, err
	}
	p := player.NewPlayerFromData(info.GetUsername(), playerUUID, info.GetXuid(), location, s.GetGamemode(), s.GetForceGamemode(), info.GetSkin(), offlinePlayerData, s.worldManager.GetWorldByName)
	p.SetServer(s)
	if offlinePlayerData == nil {
		p.OnGround = true //TODO: this hack is needed for new players in-air ticks - they don't get detected as on-ground until they move
	}
	return p, nil
}

// AddOnlinePlayer is a port of Server::addOnlinePlayer (without the cancellable PlayerLoginEvent).
func (s *Server) AddOnlinePlayer(p *player.Player) bool {
	session, _ := p.GetNetworkSession().(*mcpe.NetworkSession)
	pos := p.GetPosition()
	if session != nil {
		// KnownTranslationFactory::pocketmine_player_logIn, English text.
		s.logger.Info(fmt.Sprintf("%s%s%s[/%s:%d] logged in with entity id %d at (%s, %v, %v, %v)",
			utils.Aqua, p.GetName(), utils.Reset, session.GetIp(), session.GetPort(), p.GetID(),
			p.GetWorld().GetDisplayName(), round4(pos.X), round4(pos.Y), round4(pos.Z)))
	}

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
	for i, other := range s.playerList {
		if other == p {
			s.playerList = append(s.playerList[:i], s.playerList[i+1:]...)
			if session, ok := p.GetNetworkSession().(*mcpe.NetworkSession); ok {
				// KnownTranslationFactory::pocketmine_player_logOut, English text.
				s.logger.Info(fmt.Sprintf("%s%s%s[/%s:%d] logged out", utils.Aqua, p.GetName(), utils.Reset, session.GetIp(), session.GetPort()))
			}
			for _, remaining := range s.playerList {
				if session, ok := remaining.GetNetworkSession().(*mcpe.NetworkSession); ok {
					session.OnPlayerRemoved(p)
				}
			}
			return
		}
	}
}

// GetOnlinePlayers is a port of Server::getOnlinePlayers.
func (s *Server) GetOnlinePlayers() []*player.Player {
	return append([]*player.Player(nil), s.playerList...)
}

// BroadcastMessage is a port of Server::broadcastMessage. Broadcast channels aren't ported, so nil
// recipients means every online player (the BROADCAST_CHANNEL_USERS subscribers) plus the console.
func (s *Server) BroadcastMessage(message any, recipients []*player.Player) int {
	if recipients == nil {
		recipients = s.GetOnlinePlayers()
		s.logger.Info(consoleText(message))
	}
	for _, p := range recipients {
		p.SendMessage(message)
	}
	return len(recipients)
}

// DispatchCommand is a port of Server::dispatchCommand for players. The command map isn't wired to
// players yet (players aren't command senders), so every command is unknown:
// KnownTranslationFactory::commands_generic_notFound, prefixed red.
func (s *Server) DispatchCommand(sender *player.Player, commandLine string) bool {
	sender.SendMessage(lang.NewTranslatable("commands.generic.notFound", nil).Prefix(utils.Red))
	return false
}

// consoleText renders a message for the console log. Without a Language, a Translatable is shown
// as its key and parameters.
func consoleText(message any) string {
	if t, ok := message.(*lang.Translatable); ok {
		params := make([]string, 0, len(t.Parameters()))
		for _, p := range t.Parameters() {
			params = append(params, fmt.Sprint(p))
		}
		return strings.TrimSpace(t.Text() + " " + strings.Join(params, " "))
	}
	return fmt.Sprint(message)
}

// round4 is PHP's round($x, 4).
func round4(x float64) float64 { return stdmath.Round(x*10000) / 10000 }
