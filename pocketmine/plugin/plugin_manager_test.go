package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/event"
	pluginevent "pocketmine-go/pocketmine/event/plugin"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/scheduler"
)

// recordingLogger keeps the log lines.
type recordingLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *recordingLogger) Emergency(m string) { l.Log(log.Emergency, m) }
func (l *recordingLogger) Alert(m string)     { l.Log(log.Alert, m) }
func (l *recordingLogger) Critical(m string)  { l.Log(log.Critical, m) }
func (l *recordingLogger) Error(m string)     { l.Log(log.Error, m) }
func (l *recordingLogger) Warning(m string)   { l.Log(log.Warning, m) }
func (l *recordingLogger) Notice(m string)    { l.Log(log.Notice, m) }
func (l *recordingLogger) Info(m string)      { l.Log(log.Info, m) }
func (l *recordingLogger) Debug(m string)     { l.Log(log.Debug, m) }
func (l *recordingLogger) Log(level log.Level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, message)
}
func (l *recordingLogger) LogException(err error, trace string) { l.Log(log.Error, err.Error()) }
func (l *recordingLogger) contains(s string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, line := range l.lines {
		if strings.Contains(line, s) {
			return true
		}
	}
	return false
}

type testServer struct {
	logger     *recordingLogger
	language   *lang.Language
	commandMap *command.SimpleCommandMap
}

func (s *testServer) GetLogger() log.Logger                          { return s.logger }
func (s *testServer) GetLanguage() *lang.Language                    { return s.language }
func (s *testServer) GetSimpleCommandMap() *command.SimpleCommandMap { return s.commandMap }
func (s *testServer) GetApiVersion() string                          { return "5.44.4" }
func (s *testServer) GetBroadcastChannelSubscribers(string) []command.Sender {
	return nil
}
func (s *testServer) GetCommandMap() command.CommandMap      { return s.commandMap }
func (s *testServer) GetCommandAliases() map[string][]string { return nil }

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	language, err := lang.NewLanguage("eng", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if permission.GetManager().GetPermission(permission.RootUser) == nil {
		permission.RegisterCorePermissions()
	}
	s := &testServer{logger: &recordingLogger{}, language: language}
	s.commandMap = command.NewSimpleCommandMap(s)
	return s
}

// testLoader loads in-memory plugins: file name => plugin manifest (YAML).
type testLoader struct {
	manifests map[string]string
	created   map[string]*testPlugin
	events    *[]string
}

func (l *testLoader) CanLoadPlugin(path string) bool {
	_, ok := l.manifests[filepath.Base(path)]
	return ok
}
func (l *testLoader) LoadPlugin(string) {}
func (l *testLoader) GetPluginDescription(file string) (*Description, error) {
	return NewDescriptionFromYAML(l.manifests[filepath.Base(file)])
}
func (l *testLoader) GetAccessProtocol() string { return "" }
func (l *testLoader) GetPluginFiles(dir string) []string {
	var files []string
	for name := range l.manifests {
		files = append(files, filepath.Join(dir, name))
	}
	return files
}
func (l *testLoader) NewPlugin(file string, description *Description) (Plugin, bool) {
	p := &testPlugin{events: l.events}
	if l.created == nil {
		l.created = map[string]*testPlugin{}
	}
	l.created[description.GetName()] = p
	return p, true
}
func (l *testLoader) GetResourceProvider(string) ResourceProvider {
	return NewFSResourceProvider(fstest.MapFS{"config.toml": {Data: []byte("greeting = \"hello\"\n")}})
}

type testPlugin struct {
	PluginBase
	events      *[]string
	failEnable  bool
	commandArgs []string
}

func (p *testPlugin) OnLoad() { *p.events = append(*p.events, "load "+p.GetName()) }
func (p *testPlugin) OnEnable() error {
	*p.events = append(*p.events, "enable "+p.GetName())
	if p.failEnable {
		return &DisablePluginException{}
	}
	return nil
}
func (p *testPlugin) OnDisable() { *p.events = append(*p.events, "disable "+p.GetName()) }
func (p *testPlugin) OnCommand(sender command.Sender, cmd command.CommandLike, label string, args []string) bool {
	p.commandArgs = args
	return true
}

func manifest(name string, extra string) string {
	return fmt.Sprintf("name: %s\nversion: 1.0.0\nmain: test\\%s\napi: [5.0.0]\n%s", name, name, extra)
}

func newTestManager(t *testing.T, manifests map[string]string) (*PluginManager, *testServer, *testLoader, *[]string) {
	t.Helper()
	server := newTestServer(t)
	m, err := NewPluginManager(server, filepath.Join(t.TempDir(), "plugin_data"), nil)
	if err != nil {
		t.Fatal(err)
	}
	events := &[]string{}
	loader := &testLoader{manifests: manifests, events: events}
	m.RegisterInterface(loader)
	return m, server, loader, events
}

func TestLoadPluginsResolvesDependencies(t *testing.T) {
	m, _, _, events := newTestManager(t, map[string]string{
		"a": manifest("Alpha", "depend: [Beta]\n"),
		"b": manifest("Beta", ""),
		"c": manifest("Gamma", "softdepend: [Alpha, Missing]\n"),
		"d": manifest("Delta", "loadbefore: [Beta]\n"),
	})
	errors := 0
	loaded := m.LoadPlugins(t.TempDir(), &errors)
	if errors != 0 || len(loaded) != 4 {
		t.Fatalf("loaded %d plugins with %d errors", len(loaded), errors)
	}
	index := func(e string) int { return slices.Index(*events, e) }
	if !(index("load Delta") < index("load Beta") && index("load Beta") < index("load Alpha") && index("load Alpha") < index("load Gamma")) {
		t.Errorf("load order = %v, want Delta < Beta < Alpha < Gamma", *events)
	}
}

func TestLoadPluginsReportsBadDependencies(t *testing.T) {
	m, server, _, _ := newTestManager(t, map[string]string{
		"a": manifest("Alpha", "depend: [Nope]\n"),
		"b": manifest("Beta", "depend: [Gamma]\n"),
		"c": manifest("Gamma", "depend: [Beta]\n"),
		"d": manifest("PocketMineHelper", ""),
		"e": strings.Replace(manifest("Future", ""), "[5.0.0]", "[6.0.0]", 1),
	})
	errors := 0
	m.LoadPlugins(t.TempDir(), &errors)
	if errors != 5 {
		t.Errorf("load errors = %d, want 5", errors)
	}
	for _, want := range []string{"Unknown dependency", "Circular dependency", "Restricted name", "Incompatible API"} {
		if !server.logger.contains(want) {
			t.Errorf("log is missing %q: %v", want, server.logger.lines)
		}
	}
}

func TestEnableAndDisablePlugins(t *testing.T) {
	m, server, loader, events := newTestManager(t, map[string]string{
		"a": manifest("Alpha", "depend: [Beta]\ncommands:\n  hello:\n    description: Says hello\n    aliases: [hi]\n    permission: alpha.hello\npermissions:\n  alpha.hello:\n    default: true\n"),
		"b": manifest("Beta", ""),
	})
	t.Cleanup(func() { permission.GetManager().RemovePermission("alpha.hello") })
	m.LoadPlugins(t.TempDir(), nil)

	var enabledEvents []string
	handle := event.RegisterListener(event.Global(), nil, event.Normal, false, func(e *pluginevent.PluginEnableEvent) {
		enabledEvents = append(enabledEvents, e.GetPlugin().GetName())
	})
	defer handle.Unregister()

	for _, p := range m.GetPlugins() {
		if !m.EnablePlugin(p) {
			t.Fatalf("enabling %s failed", p.GetName())
		}
	}
	if len(enabledEvents) != 2 {
		t.Errorf("PluginEnableEvent fired for %v", enabledEvents)
	}

	// The manifest's command is registered and routed to the plugin's OnCommand.
	cmd := server.commandMap.GetCommand("hi")
	if cmd == nil {
		t.Fatal("the plugin's command alias isn't registered")
	}
	if cmd.(*command.PluginCommand).GetOwningPlugin() != loader.created["Alpha"] {
		t.Error("the command isn't owned by the plugin")
	}
	if _, err := cmd.Execute(nil, "hi", []string{"there"}); err != nil || !slices.Equal(loader.created["Alpha"].commandArgs, []string{"there"}) {
		t.Errorf("OnCommand got %v (%v)", loader.created["Alpha"].commandArgs, err)
	}
	if loader.created["Alpha"].GetCommand("hello") == nil {
		t.Error("GetCommand didn't find the plugin's own command")
	}

	// Beta has a dependent (Alpha), so Alpha is disabled first.
	*events = nil
	m.DisablePlugins()
	if !slices.Equal(*events, []string{"disable Alpha", "disable Beta"}) {
		t.Errorf("disable order = %v", *events)
	}
}

func TestPluginDisablingItselfOnEnable(t *testing.T) {
	m, server, loader, _ := newTestManager(t, map[string]string{"a": manifest("Alpha", "")})
	m.LoadPlugins(t.TempDir(), nil)
	loader.created["Alpha"].failEnable = true
	if m.EnablePlugin(loader.created["Alpha"]) {
		t.Error("EnablePlugin succeeded for a plugin that disabled itself")
	}
	if !server.logger.contains("disabled itself") && !server.logger.contains("Alpha") {
		t.Errorf("no error logged: %v", server.logger.lines)
	}
}

type testListener struct {
	calls *[]string
}

func (l *testListener) OnEnable(e *pluginevent.PluginEnableEvent) {
	*l.calls = append(*l.calls, "enable")
}
func (l *testListener) OnDisable(e *pluginevent.PluginDisableEvent) {
	*l.calls = append(*l.calls, "disable")
}
func (l *testListener) Helper(e *pluginevent.PluginDisableEvent) {
	*l.calls = append(*l.calls, "helper")
}
func (l *testListener) EventHandlerTags() map[string]map[string]string {
	return map[string]map[string]string{"Helper": {event.ListenerTagNotHandler: ""}, "OnDisable": {event.ListenerTagPriority: "MONITOR"}}
}

func TestRegisterEventsAndSchedulers(t *testing.T) {
	m, _, loader, _ := newTestManager(t, map[string]string{"a": manifest("Alpha", ""), "b": manifest("Beta", "")})
	m.LoadPlugins(t.TempDir(), nil)
	alpha, beta := loader.created["Alpha"], loader.created["Beta"]

	calls := &[]string{}
	if err := m.RegisterEvents(&testListener{calls: calls}, alpha); err == nil {
		t.Error("registering events for a disabled plugin succeeded")
	}
	m.EnablePlugin(alpha)
	if err := m.RegisterEvents(&testListener{calls: calls}, alpha); err != nil {
		t.Fatal(err)
	}

	ran := 0
	if _, err := alpha.GetScheduler().ScheduleRepeatingTask(scheduler.NewClosureTask(func() { ran++ }), 1); err != nil {
		t.Fatal(err)
	}
	m.TickSchedulers(1)
	m.TickSchedulers(2)
	if ran != 2 {
		t.Errorf("task ran %d times, want 2", ran)
	}

	m.EnablePlugin(beta)
	m.DisablePlugin(beta)
	if !slices.Equal(*calls, []string{"enable", "disable"}) {
		t.Errorf("listener calls = %v (Helper must be skipped)", *calls)
	}

	// Disabling Alpha removes its listeners and stops its scheduler.
	m.DisablePlugin(alpha)
	*calls = nil
	m.EnablePlugin(beta)
	if len(*calls) != 0 {
		t.Errorf("listener still called after its plugin was disabled: %v", *calls)
	}
	m.TickSchedulers(3)
	if ran != 2 {
		t.Error("a disabled plugin's task still ran")
	}
}

func TestPluginBaseConfig(t *testing.T) {
	m, _, loader, _ := newTestManager(t, map[string]string{"a": manifest("Alpha", "")})
	m.LoadPlugins(t.TempDir(), nil)
	alpha := loader.created["Alpha"]
	if got := alpha.GetConfig().Get("greeting", nil); got != "hello" {
		t.Errorf("config greeting = %v, want the default config.toml's", got)
	}
	if _, err := os.Stat(filepath.Join(alpha.GetDataFolder(), "config.toml")); err != nil {
		t.Errorf("config.toml wasn't saved to the data folder: %v", err)
	}
	if !strings.HasSuffix(alpha.GetDataFolder(), filepath.Join("plugin_data", "Alpha")+"/") {
		t.Errorf("data folder = %q", alpha.GetDataFolder())
	}
}

func TestPluginGraylist(t *testing.T) {
	g, err := PluginGraylistFromArray(map[string]any{"mode": "whitelist", "plugins": []any{"Alpha", 5}})
	if err != nil {
		t.Fatal(err)
	}
	if !g.IsAllowed("Alpha") || !g.IsAllowed("5") || g.IsAllowed("Beta") {
		t.Error("whitelist doesn't allow exactly its plugins")
	}
	if _, err := PluginGraylistFromArray(map[string]any{"plugins": []any{}}); err == nil {
		t.Error("missing mode accepted")
	}

	m, server, _, _ := newTestManager(t, map[string]string{"a": manifest("Alpha", ""), "b": manifest("Beta", "")})
	m.graylist = NewPluginGraylist([]string{"Beta"}, false)
	errors := 0
	loaded := m.LoadPlugins(t.TempDir(), &errors)
	if errors != 0 || len(loaded) != 1 || loaded["Alpha"] == nil {
		t.Errorf("blacklist: loaded %v, %d errors", loaded, errors)
	}
	if !server.logger.contains("blacklist") {
		t.Error("blacklisted plugin not reported")
	}
}

// A config.yml left in the data folder by an older version becomes config.toml with its settings.
func TestPluginBaseConvertsOldConfig(t *testing.T) {
	dataPath := t.TempDir()
	m, _, loader, _ := newTestManager(t, map[string]string{"a": manifest("Alpha", "")})
	m.LoadPlugins(dataPath, nil)
	alpha := loader.created["Alpha"]
	old := filepath.Join(alpha.GetDataFolder(), "config.yml")
	if err := os.WriteFile(old, []byte("greeting: changed by the owner\nnested:\n  level: 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	alpha.ReloadConfig()
	if got := alpha.GetConfig().Get("greeting", nil); got != "changed by the owner" {
		t.Errorf("greeting = %v, want the owner's value from config.yml", got)
	}
	if got := alpha.GetConfig().GetNested("nested.level", nil); got != 3 {
		t.Errorf("nested.level = %v (%T), want 3", got, got)
	}
	if _, err := os.Stat(old + ".bak"); err != nil {
		t.Errorf("config.yml wasn't kept as config.yml.bak: %v", err)
	}
}
