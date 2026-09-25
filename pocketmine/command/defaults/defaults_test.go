package defaults

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pocketmine-go/pocketmine/console"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/server"
	"pocketmine-go/pocketmine/world"
)

// recordingConsole is a ConsoleCommandSender that keeps what it's sent.
type recordingConsole struct {
	*console.ConsoleCommandSender
	messages []string
}

func (r *recordingConsole) SendMessage(message any) {
	if t, ok := message.(*lang.Translatable); ok {
		r.messages = append(r.messages, r.GetLanguage().Translate(t))
		return
	}
	r.messages = append(r.messages, message.(string))
}

func newTestServer(t *testing.T) (*server.Server, *recordingConsole) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte("level-seed=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := server.New(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	return s, &recordingConsole{ConsoleCommandSender: console.NewConsoleCommandSender(s, s.GetLanguage())}
}

func TestDefaultCommandsAreRegistered(t *testing.T) {
	s, _ := newTestServer(t)
	for _, name := range []string{"stop", "help", "?", "list", "op", "deop", "ban", "ban-ip", "banlist", "pardon", "unban", "whitelist", "time", "tp", "gamemode", "kill", "xp", "version", "status", "save-all", "seed", "difficulty"} {
		if s.GetSimpleCommandMap().GetCommand(name) == nil {
			t.Errorf("command %q isn't registered", name)
		}
	}
}

func TestConsoleCommands(t *testing.T) {
	s, sender := newTestServer(t)

	s.DispatchCommand(sender, "op Alex", false)
	if !s.IsOp("alex") {
		t.Error("/op didn't add the op")
	}
	s.DispatchCommand(sender, "deop Alex", false)
	if s.IsOp("alex") {
		t.Error("/deop didn't remove the op")
	}

	s.DispatchCommand(sender, "whitelist on", false)
	s.DispatchCommand(sender, "whitelist add Steve", false)
	if !s.HasWhitelist() || !s.IsWhitelisted("steve") || s.IsWhitelisted("herobrine") {
		t.Error("/whitelist on/add didn't work")
	}

	s.DispatchCommand(sender, "ban Griefer being mean", false)
	if !s.GetNameBans().IsBanned("griefer") {
		t.Error("/ban didn't ban")
	}
	s.DispatchCommand(sender, "pardon Griefer", false)
	if s.GetNameBans().IsBanned("griefer") {
		t.Error("/pardon didn't unban")
	}

	s.DispatchCommand(sender, "time set noon", false)
	if got := s.GetWorldManager().GetDefaultWorld().GetTime(); got != world.TimeNoon {
		t.Errorf("/time set noon: time = %d", got)
	}

	s.DispatchCommand(sender, "difficulty hard", false)
	if got := s.GetWorldManager().GetDefaultWorld().GetDifficulty(); got != world.DifficultyHard {
		t.Errorf("/difficulty hard: difficulty = %d", got)
	}

	sender.messages = nil
	s.DispatchCommand(sender, "list", false)
	if len(sender.messages) != 2 || !strings.Contains(sender.messages[0], "0/20") {
		t.Errorf("/list output = %q", sender.messages)
	}

	sender.messages = nil
	s.DispatchCommand(sender, "gamemode creative", false)
	if len(sender.messages) == 0 {
		t.Error("/gamemode from the console should print its usage (no player target)")
	}

	s.DispatchCommand(sender, "stop", false)
	if s.IsRunning() {
		t.Error("/stop didn't stop the server")
	}
}

func TestHelpListsCommands(t *testing.T) {
	s, sender := newTestServer(t)
	s.DispatchCommand(sender, "help", false)
	joined := strings.Join(sender.messages, "\n")
	if !strings.Contains(joined, "/ban:") {
		t.Errorf("/help output doesn't list /ban:\n%s", joined)
	}
	sender.messages = nil
	s.DispatchCommand(sender, "help stop", false)
	if !strings.Contains(strings.Join(sender.messages, "\n"), "stop") {
		t.Errorf("/help stop output = %q", sender.messages)
	}
}

func TestPhpCasts(t *testing.T) {
	if phpInt("12abc") != 12 || phpInt("abc") != 0 || phpInt("-5") != -5 {
		t.Error("phpInt")
	}
	if phpFloat("1.5x") != 1.5 || phpFloat("") != 0 {
		t.Error("phpFloat")
	}
	if getRelativeDouble(10, "~5", MinCoord, MaxCoord) != 15 || getRelativeDouble(10, "3", MinCoord, MaxCoord) != 3 {
		t.Error("getRelativeDouble")
	}
	if numberFormat(1234567) != "1,234,567" || numberFormat(12) != "12" {
		t.Error("numberFormat")
	}
}
