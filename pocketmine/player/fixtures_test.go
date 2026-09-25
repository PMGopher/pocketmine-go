package player

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/form"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

// fakeServer records what Player asks of the server.
type fakeServer struct {
	online       []*Player
	broadcast    []any
	commands     []string
	saved        map[string]*nbt.CompoundTag
	removed      []*Player
	tick         int64
	gamemode     GameMode
	force        bool
	worldManager *world.WorldManager
	bans         *permission.BanList
	subscribers  map[string][]command.Sender
}

var testLanguage = func() *lang.Language {
	l, err := lang.NewLanguage("eng", "", "")
	if err != nil {
		panic(err)
	}
	return l
}()

func (s *fakeServer) GetBroadcastChannelSubscribers(channel string) []command.Sender {
	return s.subscribers[channel]
}
func (s *fakeServer) GetCommandMap() command.CommandMap       { return nil }
func (s *fakeServer) GetCommandAliases() map[string][]string  { return nil }
func (s *fakeServer) GetLanguage() *lang.Language             { return testLanguage }
func (s *fakeServer) GetLogger() log.Logger                   { return log.NewSimpleLogger() }
func (s *fakeServer) GetTick() int64                          { return s.tick }
func (s *fakeServer) GetOnlinePlayers() []*Player             { return s.online }
func (s *fakeServer) GetWorldManager() *world.WorldManager    { return s.worldManager }
func (s *fakeServer) IsOp(name string) bool                   { return false }
func (s *fakeServer) GetNameBans() *permission.BanList        { return s.bans }
func (s *fakeServer) IsHardcore() bool                        { return false }
func (s *fakeServer) GetGamemode() GameMode                   { return s.gamemode }
func (s *fakeServer) GetForceGamemode() bool                  { return s.force }
func (s *fakeServer) GetAllowedViewDistance(distance int) int { return max(2, min(distance, 16)) }
func (s *fakeServer) GetPropertyInt(v string, def int) int    { return def }
func (s *fakeServer) GetConfigBool(v string, def bool) bool   { return def }
func (s *fakeServer) BroadcastMessage(message any, recipients []command.Sender) int {
	s.broadcast = append(s.broadcast, message)
	if recipients == nil {
		recipients = s.GetBroadcastChannelSubscribers(BroadcastChannelUsers)
	}
	for _, p := range recipients {
		p.SendMessage(message)
	}
	return len(recipients)
}
func (s *fakeServer) DispatchCommand(sender command.Sender, commandLine string, internal bool) bool {
	s.commands = append(s.commands, commandLine)
	return true
}
func (s *fakeServer) SubscribeToBroadcastChannel(channelID string, subscriber command.Sender) {
	s.UnsubscribeFromBroadcastChannel(channelID, subscriber)
	if s.subscribers == nil {
		s.subscribers = map[string][]command.Sender{}
	}
	s.subscribers[channelID] = append(s.subscribers[channelID], subscriber)
}
func (s *fakeServer) UnsubscribeFromBroadcastChannel(channelID string, subscriber command.Sender) {
	list := s.subscribers[channelID]
	for i, sub := range list {
		if sub == subscriber {
			s.subscribers[channelID] = append(list[:i:i], list[i+1:]...)
			return
		}
	}
}
func (s *fakeServer) UnsubscribeFromAllBroadcastChannels(subscriber command.Sender) {
	for channel := range s.subscribers {
		s.UnsubscribeFromBroadcastChannel(channel, subscriber)
	}
}
func (s *fakeServer) RemoveOnlinePlayer(p *Player) {
	s.removed = append(s.removed, p)
	for i, o := range s.online {
		if o == p {
			s.online = append(s.online[:i], s.online[i+1:]...)
		}
	}
}
func (s *fakeServer) SaveOfflinePlayerData(name string, data *nbt.CompoundTag) {
	if s.saved == nil {
		s.saved = map[string]*nbt.CompoundTag{}
	}
	s.saved[name] = data
}

// fakeSession records what Player sends to its client.
type fakeSession struct {
	disconnected    bool
	packets         []packet.Packet
	messages        []any
	viewAreaSyncs   int
	startedChunks   [][2]int
	pendingChunks   []func()
	abilitySyncs    int
	gameModeSyncs   []GameMode
	movementSyncs   int
	deaths, respawn int
	destroyed       bool
}

func (s *fakeSession) SendDataPacket(pk packet.Packet)                    { s.packets = append(s.packets, pk) }
func (s *fakeSession) IsConnected() bool                                  { return !s.disconnected }
func (s *fakeSession) GetIp() string                                      { return "127.0.0.1" }
func (s *fakeSession) GetPort() int                                       { return 19132 }
func (s *fakeSession) OnChatMessage(message any)                          { s.messages = append(s.messages, message) }
func (s *fakeSession) OnJukeboxPopup(message any)                         {}
func (s *fakeSession) OnPopup(message string)                             {}
func (s *fakeSession) OnTip(message string)                               {}
func (s *fakeSession) OnTitle(title string)                               {}
func (s *fakeSession) OnSubTitle(subtitle string)                         {}
func (s *fakeSession) OnActionBar(actionBar string)                       {}
func (s *fakeSession) OnClearTitle()                                      {}
func (s *fakeSession) OnResetTitleOptions()                               {}
func (s *fakeSession) OnTitleDuration(fadeIn, stay, fadeOut int)          {}
func (s *fakeSession) OnToastNotification(title, body string)             {}
func (s *fakeSession) OnFormSent(id int, f form.Form) bool                { return true }
func (s *fakeSession) OnCloseAllForms()                                   {}
func (s *fakeSession) OnOpenSignEditor(pos math.Vector3, frontSide bool)  {}
func (s *fakeSession) OnItemCooldownChanged(it item.Item, ticks int)      {}
func (s *fakeSession) SyncViewAreaRadius(distance int)                    {}
func (s *fakeSession) SyncViewAreaCenterPoint(pos math.Vector3, dist int) { s.viewAreaSyncs++ }
func (s *fakeSession) SyncPlayerSpawnPoint(newSpawn math.Vector3)         {}
func (s *fakeSession) SyncGameMode(mode GameMode, isRollback bool) {
	s.gameModeSyncs = append(s.gameModeSyncs, mode)
}
func (s *fakeSession) SyncAbilities(p *Player) { s.abilitySyncs++ }
func (s *fakeSession) SyncAdventureSettings()  {}
func (s *fakeSession) SyncMovement(pos math.Vector3, yaw, pitch *float64, mode byte) {
	s.movementSyncs++
}
func (s *fakeSession) OnEnterWorld() {}
func (s *fakeSession) StartUsingChunk(chunkX, chunkZ int, onCompletion func()) {
	s.startedChunks = append(s.startedChunks, [2]int{chunkX, chunkZ})
	s.pendingChunks = append(s.pendingChunks, onCompletion)
}
func (s *fakeSession) StopUsingChunk(chunkX, chunkZ int) {}
func (s *fakeSession) NotifyTerrainReady()               {}
func (s *fakeSession) OnServerDeath(deathMessage any)    { s.deaths++ }
func (s *fakeSession) OnServerRespawn()                  { s.respawn++ }
func (s *fakeSession) OnPlayerDestroyed(reason, screen any) {
	s.destroyed = true
	s.disconnected = true
}
func (s *fakeSession) Transfer(ip string, port int, r any)    { s.disconnected = true }
func (s *fakeSession) DisconnectWithError(reason, screen any) { s.disconnected = true }
func (s *fakeSession) GetInvManager() InventoryManager        { return nil }

// completeChunkSends runs the completion callbacks of the chunks sent so far (the network
// finishing sending them).
func (s *fakeSession) completeChunkSends() {
	pending := s.pendingChunks
	s.pendingChunks = nil
	for _, cb := range pending {
		cb()
	}
}

func init() {
	permission.RegisterCorePermissions()
}

// testWorldManagers maps each test world to the manager it's the default world of.
var testWorldManagers = map[*world.World]*world.WorldManager{}

// newTestWorld generates a flat world in a temporary folder, as the default world of its own
// WorldManager (PHP's Player::getSpawn falls back to the default world).
func newTestWorld(t *testing.T) *world.World {
	t.Helper()
	m := world.NewWorldManager(t.TempDir(), convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
	})
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	options := world.NewWorldCreationOptions()
	options.GeneratorName = "flat"
	w, err := m.GenerateWorld("world", gen, options)
	if err != nil {
		t.Fatal(err)
	}
	m.SetDefaultWorld(w)
	testWorldManagers[w] = m
	t.Cleanup(func() {
		delete(testWorldManagers, w)
		_ = w.Close()
	})
	return w
}

func newTestSkin(t *testing.T) *entity.Skin {
	t.Helper()
	skin, err := entity.NewSkin("Standard_Custom", make([]byte, 64*64*4), nil, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return skin
}

func newTestInfo(t *testing.T) *XboxLivePlayerInfo {
	return NewXboxLivePlayerInfo("xuid-1", "Steve", uuid.NewString(), newTestSkin(t), "en_US", nil)
}

// newPlayerFor creates a connected player for server in w, from saved data tag (may be nil).
func newPlayerFor(t *testing.T, server *fakeServer, w *world.World, pos math.Vector3, tag *nbt.CompoundTag) (*Player, *fakeSession) {
	t.Helper()
	if server.worldManager == nil {
		server.worldManager = testWorldManagers[w]
	}
	if server.bans == nil {
		server.bans = permission.NewBanList(t.TempDir() + "/banned.txt")
	}
	session := &fakeSession{}
	p := NewPlayer(server, session, newTestInfo(t), true, entity.LocationFromObject(pos, w, 0, 0), tag)
	server.online = append(server.online, p)
	return p, session
}

// newTestPlayer creates a player in a fresh world. id is unused (entity IDs are allocated by the
// entity package, like PHP's Entity::nextRuntimeId) but kept so call sites stay readable.
func newTestPlayer(t *testing.T, id int, pos math.Vector3) *Player {
	t.Helper()
	return newTestPlayerIn(t, newTestWorld(t), pos)
}

func newTestPlayerIn(t *testing.T, w *world.World, pos math.Vector3) *Player {
	t.Helper()
	p, _ := newPlayerFor(t, &fakeServer{}, w, pos, nil)
	return p
}

func newConnectedPlayer(t *testing.T, server *fakeServer) (*Player, *fakeSession) {
	t.Helper()
	return newPlayerFor(t, server, newTestWorld(t), math.NewVector3(0.5, 70, 0.5), nil)
}
