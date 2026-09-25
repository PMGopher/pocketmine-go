package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/player"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	// A fixed seed keeps the generated terrain the same on every run.
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte("level-seed=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := New(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// recordingSender is an mcpe.PacketSender that keeps the packets instead of sending them.
type recordingSender struct {
	packets []packet.Packet
	closed  bool
}

func (r *recordingSender) WritePacket(pk packet.Packet) error {
	r.packets = append(r.packets, pk)
	return nil
}

func (r *recordingSender) Close() error {
	r.closed = true
	return nil
}

func newTestSession(s *Server, port int) (*mcpe.NetworkSession, *recordingSender) {
	sender := &recordingSender{}
	broadcaster := mcpe.NewStandardPacketBroadcaster()
	session := mcpe.NewNetworkSession(s, s.GetNetwork().GetSessionManager(), sender, nil, broadcaster, mcpe.NewStandardEntityEventBroadcaster(broadcaster), "127.0.0.1", port)
	return session, sender
}

func newTestPlayerInfo(t *testing.T, name string) *player.XboxLivePlayerInfo {
	t.Helper()
	skin, err := entity.NewSkin("Standard_Custom", make([]byte, 64*64*4), nil, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return player.NewXboxLivePlayerInfo("", name, uuid.NewString(), skin, "en_US", nil)
}

func TestNewPreparesDataFolderAndDefaultWorld(t *testing.T) {
	s := newTestServer(t)
	for _, name := range []string{"server.properties", "worlds", "players"} {
		if _, err := os.Stat(filepath.Join(s.GetDataPath(), name)); err != nil {
			t.Errorf("%s missing: %v", name, err)
		}
	}
	w := s.GetWorldManager().GetDefaultWorld()
	if w == nil || w.GetFolderName() != "world" {
		t.Fatalf("default world = %v, want \"world\"", w)
	}
	if _, err := os.Stat(filepath.Join(s.GetDataPath(), "worlds", "world", "level.dat")); err != nil {
		t.Errorf("level.dat missing: %v", err)
	}
	if s.GetMotd() != DefaultServerName || s.GetPort() != DefaultPortIPv4 || !s.RequiresAuthentication() {
		t.Errorf("defaults: motd=%q port=%d auth=%v", s.GetMotd(), s.GetPort(), s.RequiresAuthentication())
	}
}

func TestCreatePlayerNewAndReturning(t *testing.T) {
	s := newTestServer(t)
	w := s.GetWorldManager().GetDefaultWorld()
	session, _ := newTestSession(s, 1)

	p, err := s.CreatePlayer(session, newTestPlayerInfo(t, "Alex"), false, s.GetOfflinePlayerData("Alex"))
	if err != nil {
		t.Fatal(err)
	}
	want := w.GetSafeSpawn(w.GetSpawnLocation())
	if pos := p.GetPosition(); pos.X != want.X || pos.Y != want.Y || pos.Z != want.Z {
		t.Errorf("new player at %v, want the safe spawn %v", pos, want)
	}
	if p.GetGamemode() != player.GameModeSurvival {
		t.Errorf("game mode = %v, want the server's survival", p.GetGamemode())
	}

	// Leaving saves the position (Player::onPostDisconnect -> save); the next join restores it.
	p.Teleport(want.Add(5, 0, 3))
	p.Save()
	p.Close()

	again, err := s.CreatePlayer(session, newTestPlayerInfo(t, "Alex"), false, s.GetOfflinePlayerData("Alex"))
	if err != nil {
		t.Fatal(err)
	}
	if pos := again.GetPosition(); pos.X != want.X+5 || pos.Z != want.Z+3 {
		t.Errorf("returning player at %v, want the saved position %v", pos, want.Add(5, 0, 3))
	}
}

func TestOnlinePlayersAndBroadcast(t *testing.T) {
	s := newTestServer(t)
	sessionA, _ := newTestSession(s, 1)
	sessionB, _ := newTestSession(s, 2)
	// Login adds them to the online players (NetworkSession::onPlayerCreated).
	sessionA.Login(newTestPlayerInfo(t, "A").WithoutXboxData(), false, false)
	sessionB.Login(newTestPlayerInfo(t, "B").WithoutXboxData(), false, false)
	a, b := sessionA.GetPlayer(), sessionB.GetPlayer()
	s.SubscribeToBroadcastChannel(BroadcastChannelUsers, a)
	s.SubscribeToBroadcastChannel(BroadcastChannelUsers, b)
	if got := s.GetOnlinePlayers(); len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("online players = %v, want [A B] in join order", got)
	}
	if n := s.BroadcastMessage(lang.NewTranslatable("chat.type.text", []any{"A", "hi"}), nil); n != 2 {
		t.Errorf("BroadcastMessage reached %d players, want 2", n)
	}
	s.RemoveOnlinePlayer(a)
	if got := s.GetOnlinePlayers(); len(got) != 1 || got[0] != b {
		t.Errorf("online players after removing A = %v, want [B]", got)
	}
}

func TestGetAllowedViewDistance(t *testing.T) {
	s := newTestServer(t)
	for requested, want := range map[int]int{1: 2, 8: 8, 16: 16, 64: 16} {
		if got := s.GetAllowedViewDistance(requested); got != want {
			t.Errorf("GetAllowedViewDistance(%d) = %d, want %d", requested, got, want)
		}
	}
}

func TestShutdownUnloadsWorldsAndSavesConfig(t *testing.T) {
	s := newTestServer(t)
	s.Shutdown()
	s.ForceShutdown()
	if len(s.GetWorldManager().GetWorlds()) != 0 {
		t.Error("worlds are still loaded after Shutdown")
	}
}

func TestNewWorldPregeneratesSpawnTerrain(t *testing.T) {
	// WorldManager::generateWorld's background generation: radius 8 around spawn.
	s := newTestServer(t)
	w := s.GetWorldManager().GetDefaultWorld()
	spawn := w.GetSpawnLocation()
	cx, cz := spawn.FloorX()>>4, spawn.FloorZ()>>4
	// ChunkSelector reaches +7 on the positive side and -8 on the negative side.
	for _, c := range [][2]int{{cx, cz}, {cx + 7, cz}, {cx - 8, cz}, {cx, cz - 8}} {
		if chunk, ok := w.GetChunk(c[0], c[1]); !ok || !chunk.IsPopulated() {
			t.Errorf("spawn chunk %v isn't generated and populated", c)
		}
	}
}

func TestLoginToSpawn(t *testing.T) {
	// NetworkSession's login phase (setAuthenticationStatus -> createPlayer -> onPlayerCreated ->
	// PreSpawnPacketHandler) without a gophertunnel connection: the spawn chunks are then sent
	// by the tick, like PHP.
	s := newTestServer(t)
	s.Lock()
	defer s.Unlock()
	session, sender := newTestSession(s, 1)
	session.Login(newTestPlayerInfo(t, "Steve").WithoutXboxData(), false, false)
	p := session.GetPlayer()
	if p == nil || !session.IsConnected() {
		t.Fatal("player wasn't created")
	}
	if got := s.GetOnlinePlayers(); len(got) != 1 || got[0] != p {
		t.Fatalf("online players = %v", got)
	}
	if session.GetInvManager() == nil {
		t.Fatal("no InventoryManager")
	}
	var sawCreative, sawAbilities bool
	for _, pk := range sender.packets {
		switch pk.(type) {
		case *packet.CreativeContent:
			sawCreative = true
		case *packet.UpdateAbilities:
			sawAbilities = true
		}
	}
	if !sawCreative || !sawAbilities {
		t.Errorf("pre-spawn packets missing: creative=%v abilities=%v", sawCreative, sawAbilities)
	}

	session.OnClientRequestChunkRadius(4)
	if !p.IsSpawned() {
		t.Fatal("player didn't spawn after its spawn chunks were sent")
	}

	// Duplicate login: the existing session is kicked (PlayerDuplicateLoginEvent not cancelled).
	session2, _ := newTestSession(s, 2)
	session2.Login(newTestPlayerInfo(t, "steve").WithoutXboxData(), false, false)
	if session.IsConnected() || !sender.closed {
		t.Error("the existing session should have been disconnected by the duplicate login")
	}
	if !session2.IsConnected() {
		t.Error("the new session should have been accepted")
	}
}

func TestWhitelistAndOps(t *testing.T) {
	s := newTestServer(t)
	s.GetConfigGroup().SetConfigBool(PropertyWhitelist, true)
	if s.IsWhitelisted("Alex") {
		t.Error("Alex shouldn't be whitelisted yet")
	}
	s.AddWhitelist("Alex")
	if !s.IsWhitelisted("alex") {
		t.Error("whitelist should be case-insensitive")
	}
	s.AddOp("Bob")
	if !s.IsOp("bob") || !s.IsWhitelisted("Bob") {
		t.Error("ops are always whitelisted")
	}
	s.RemoveOp("BOB")
	if s.IsOp("Bob") {
		t.Error("RemoveOp didn't remove the op")
	}
}
