package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
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

	p, err := s.CreatePlayer(nil, newTestPlayerInfo(t, "Alex"))
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

	again, err := s.CreatePlayer(nil, newTestPlayerInfo(t, "Alex"))
	if err != nil {
		t.Fatal(err)
	}
	if pos := again.GetPosition(); pos.X != want.X+5 || pos.Z != want.Z+3 {
		t.Errorf("returning player at %v, want the saved position %v", pos, want.Add(5, 0, 3))
	}
}

func TestOnlinePlayersAndBroadcast(t *testing.T) {
	s := newTestServer(t)
	a, _ := s.CreatePlayer(nil, newTestPlayerInfo(t, "A"))
	b, _ := s.CreatePlayer(nil, newTestPlayerInfo(t, "B"))
	s.AddOnlinePlayer(a)
	s.AddOnlinePlayer(b)
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
	if len(s.GetWorldManager().GetWorlds()) != 0 {
		t.Error("worlds are still loaded after Shutdown")
	}
}
