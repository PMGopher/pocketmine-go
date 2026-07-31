package command

import (
	"testing"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
)

type fakeSender struct {
	*permission.Permissible
	name     string
	messages []string
}

func newFakeSender(name string, basePerms map[string]bool) *fakeSender {
	return &fakeSender{Permissible: permission.NewPermissible(basePerms), name: name}
}

func (s *fakeSender) GetLanguage() *lang.Language { return nil }
func (s *fakeSender) SendMessage(message any) {
	s.messages = append(s.messages, stringifyMessage(message))
}
func (s *fakeSender) GetServer() Server          { return testServer }
func (s *fakeSender) GetName() string            { return s.name }
func (s *fakeSender) GetScreenLineHeight() int   { return 20 }
func (s *fakeSender) SetScreenLineHeight(h *int) {}

type fakeServer struct {
	cmdMap    CommandMap
	aliases   map[string][]string
	broadcast []Sender
}

func (s *fakeServer) GetBroadcastChannelSubscribers(channel string) []Sender { return s.broadcast }
func (s *fakeServer) GetCommandMap() CommandMap                              { return s.cmdMap }
func (s *fakeServer) GetCommandAliases() map[string][]string                 { return s.aliases }
func (s *fakeServer) GetLanguage() *lang.Language                            { return nil }

var testServer = &fakeServer{}

func ensurePermission(name string) {
	if permission.GetManager().GetPermission(name) == nil {
		permission.GetManager().AddPermission(permission.NewPermission(name, "", nil))
	}
}

func TestClosureCommandRegisterAndDispatch(t *testing.T) {
	ensurePermission("test.say")
	var called bool
	var gotArgs []string
	cmd, err := NewClosureCommand("say", func(sender Sender, cmd CommandLike, label string, args []string) (any, error) {
		called = true
		gotArgs = args
		return true, nil
	}, []string{"test.say"}, "says something", nil, nil)
	if err != nil {
		t.Fatalf("NewClosureCommand() error = %v", err)
	}

	m := NewSimpleCommandMap(testServer)
	testServer.cmdMap = m
	m.Register("pocketmine", cmd, "")

	sender := newFakeSender("Steve", map[string]bool{"test.say": true})
	if !m.Dispatch(sender, `say hello world`) {
		t.Fatalf("Dispatch() = false, want true")
	}
	if !called {
		t.Fatalf("expected the command closure to be called")
	}
	if len(gotArgs) != 2 || gotArgs[0] != "hello" || gotArgs[1] != "world" {
		t.Fatalf("gotArgs = %v, want [hello world]", gotArgs)
	}
}

func TestDispatchDeniedByPermission(t *testing.T) {
	ensurePermission("test.restricted")
	var called bool
	cmd, _ := NewClosureCommand("restricted", func(sender Sender, cmd CommandLike, label string, args []string) (any, error) {
		called = true
		return true, nil
	}, []string{"test.restricted"}, "", nil, nil)

	m := NewSimpleCommandMap(testServer)
	testServer.cmdMap = m
	m.Register("pocketmine", cmd, "")

	sender := newFakeSender("Nobody", map[string]bool{"test.restricted": false})
	m.Dispatch(sender, "restricted")

	if called {
		t.Fatalf("command should not have been called without permission")
	}
	if len(sender.messages) == 0 {
		t.Fatalf("expected a permission-denied message to be sent")
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	m := NewSimpleCommandMap(testServer)
	testServer.cmdMap = m
	sender := newFakeSender("Steve", nil)

	if m.Dispatch(sender, "totallynotacommand") {
		t.Fatalf("Dispatch() = true for an unknown command, want false")
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected exactly one message, got %v", sender.messages)
	}
}

func TestDispatchInvalidSyntaxShowsUsage(t *testing.T) {
	ensurePermission("test.usage")
	cmd, _ := NewClosureCommand("usagetest", func(sender Sender, cmd CommandLike, label string, args []string) (any, error) {
		return nil, &InvalidCommandSyntaxException{}
	}, []string{"test.usage"}, "", "/usagetest <arg>", nil)

	m := NewSimpleCommandMap(testServer)
	testServer.cmdMap = m
	m.Register("pocketmine", cmd, "")

	sender := newFakeSender("Steve", map[string]bool{"test.usage": true})
	m.Dispatch(sender, "usagetest")

	if len(sender.messages) != 1 || sender.messages[0] != "Usage: /usagetest <arg>" {
		t.Fatalf("messages = %v, want [\"Usage: /usagetest <arg>\"]", sender.messages)
	}
}

func TestParseQuoteAware(t *testing.T) {
	got := ParseQuoteAware(`give "steve jobs" apple`)
	want := []string{"give", "steve jobs", "apple"}
	if len(got) != len(want) {
		t.Fatalf("ParseQuoteAware() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseQuoteAware() = %v, want %v", got, want)
		}
	}
}

func TestParseQuoteAwareEscapedQuotes(t *testing.T) {
	got := ParseQuoteAware(`say "This is a \"string containing quotes\""`)
	want := []string{"say", `This is a "string containing quotes"`}
	if len(got) != len(want) || got[1] != want[1] {
		t.Fatalf("ParseQuoteAware() = %v, want %v", got, want)
	}
}

func TestFormattedCommandAliasResolvesPlaceholders(t *testing.T) {
	ensurePermission("test.tp")
	var gotArgs []string
	tp, _ := NewClosureCommand("tp", func(sender Sender, cmd CommandLike, label string, args []string) (any, error) {
		gotArgs = args
		return true, nil
	}, []string{"test.tp"}, "", nil, nil)

	m := NewSimpleCommandMap(testServer)
	testServer.cmdMap = m
	m.Register("pocketmine", tp, "")

	alias := NewFormattedCommandAlias("home", []string{"tp $1 0 0 0"})
	// Like the real PHP FormattedCommandAlias, this has no permissions by default (Command starts
	// with an empty permission list, and TestPermissionSilent always denies an empty list) — the
	// code that installs it (SimpleCommandMap.RegisterServerAliases) would need to grant one, same
	// as here.
	alias.SetPermissions([]string{"test.tp"})
	m.mu.Lock()
	m.knownCommands["home"] = alias
	m.mu.Unlock()

	sender := newFakeSender("Steve", map[string]bool{"test.tp": true})
	if !m.Dispatch(sender, "home Steve") {
		t.Fatalf("Dispatch(home) = false, want true")
	}
	want := []string{"Steve", "0", "0", "0"}
	if len(gotArgs) != len(want) {
		t.Fatalf("gotArgs = %v, want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("gotArgs = %v, want %v", gotArgs, want)
		}
	}
}

func TestPluginCommandSkipsDisabledPlugin(t *testing.T) {
	plugin := &fakePlugin{name: "Test", enabled: false}
	var called bool
	executor := executorFunc(func(sender Sender, cmd CommandLike, label string, args []string) bool {
		called = true
		return true
	})
	cmd := NewPluginCommand("test", plugin, executor)

	result, err := cmd.Execute(newFakeSender("Steve", nil), "test", nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result != false || called {
		t.Fatalf("expected Execute() to short-circuit for a disabled plugin")
	}
}

type executorFunc func(sender Sender, cmd CommandLike, label string, args []string) bool

func (f executorFunc) OnCommand(sender Sender, cmd CommandLike, label string, args []string) bool {
	return f(sender, cmd, label, args)
}

type fakePlugin struct {
	name    string
	enabled bool
}

func (p *fakePlugin) IsEnabled() bool { return p.enabled }
func (p *fakePlugin) Name() string    { return p.name }
