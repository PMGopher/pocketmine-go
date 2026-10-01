package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/plugin"
	"pocketmine-go/pocketmine/utils"
)

// exampleGoPlugin is a compiled-in plugin like one a server owner would write.
type exampleGoPlugin struct {
	plugin.PluginBase
	enabled, disabled int
	ticks             int
}

func (p *exampleGoPlugin) OnEnable() error {
	p.enabled++
	p.SaveDefaultConfig()
	return nil
}

func (p *exampleGoPlugin) OnDisable() { p.disabled++ }

var examplePluginInstance *exampleGoPlugin

func init() {
	plugin.RegisterGoPlugin(fstest.MapFS{
		"plugin.toml":           {Data: []byte("name = \"ExamplePlugin\"\nversion = \"1.2.3\"\nmain = \"example.Main\"\napi = [\"5.0.0\"]\nload = \"POSTWORLD\"\nauthors = [\"Alex\"]\n")},
		"resources/config.toml": {Data: []byte("enabled = true\n")},
	}, func() plugin.Plugin {
		examplePluginInstance = &exampleGoPlugin{}
		return examplePluginInstance
	})
}

func TestServerLoadsGoPlugins(t *testing.T) {
	s := newTestServer(t)
	p := s.GetPluginManager().GetPlugin("ExamplePlugin")
	if p == nil {
		t.Fatal("the registered plugin wasn't loaded")
	}
	if !p.IsEnabled() || examplePluginInstance.enabled != 1 {
		t.Fatalf("plugin enabled=%v (OnEnable ran %d times)", p.IsEnabled(), examplePluginInstance.enabled)
	}
	// pocketmine.toml has plugins.legacy-data-dir = false, so the data folder is in plugin_data.
	if want := filepath.Join(s.GetDataPath(), "plugin_data", "ExamplePlugin"); !strings.HasPrefix(p.GetDataFolder(), want) {
		t.Errorf("data folder = %q, want %q", p.GetDataFolder(), want)
	}
	if _, err := os.Stat(filepath.Join(p.GetDataFolder(), "config.toml")); err != nil {
		t.Errorf("the embedded config.toml wasn't saved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.GetDataPath(), "plugin_list.toml")); err != nil {
		t.Errorf("plugin_list.toml wasn't created: %v", err)
	}
	if plugins := s.GetQueryPlugins(); len(plugins) != 1 || plugins[0].GetVersion() != "1.2.3" {
		t.Errorf("query plugins = %v", plugins)
	}

	s.ForceShutdown()
	if p.IsEnabled() || examplePluginInstance.disabled != 1 {
		t.Error("shutting down didn't disable the plugin")
	}
}

func TestPluginListBlacklist(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte("level-seed=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin_list.toml"), []byte("mode = \"blacklist\"\nplugins = [\"ExamplePlugin\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := New(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer s.ForceShutdown()
	if s.GetPluginManager().GetPlugin("ExamplePlugin") != nil {
		t.Error("a blacklisted plugin was loaded")
	}
}

// A data folder of an older version (pocketmine.yml, plugin_list.yml) keeps its settings: the
// files become pocketmine.toml and plugin_list.toml, and the old ones are kept as .bak.
func TestOldYAMLConfigsAreConverted(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"server.properties": "level-seed=1\n",
		"pocketmine.yml":    "debug:\n  level: 2\nchunk-sending:\n  per-tick: 9\naliases:\n  savestop: [save-all, stop]\nworlds:\n",
		"plugin_list.yml":   "mode: blacklist\nplugins: [ExamplePlugin]\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s, err := New(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer s.ForceShutdown()

	if got := s.configGroup.GetPropertyInt("chunk-sending.per-tick", 0); got != 9 {
		t.Errorf("chunk-sending.per-tick = %d, want the 9 from pocketmine.yml", got)
	}
	if got := s.GetCommandAliases()["savestop"]; len(got) != 2 || got[1] != "stop" {
		t.Errorf("alias savestop = %v", got)
	}
	if s.GetPluginManager().GetPlugin("ExamplePlugin") != nil {
		t.Error("the blacklist from plugin_list.yml was lost")
	}
	for _, name := range []string{"pocketmine.toml", "pocketmine.yml.bak", "plugin_list.toml", "plugin_list.yml.bak"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s missing: %v", name, err)
		}
	}
}

// The default pocketmine.toml has every setting the server reads, with PocketMine-MP's defaults.
func TestDefaultPocketmineTOML(t *testing.T) {
	data, err := utils.ParseTOML(defaultPocketmineToml)
	if err != nil {
		t.Fatal(err)
	}
	c := func(path string) any {
		var v any = data
		for _, part := range strings.Split(path, ".") {
			m, ok := v.(map[string]any)
			if !ok {
				return nil
			}
			v = m[part]
		}
		return v
	}
	for key, want := range map[string]any{
		YmlSettingsAsyncWorkers:               "auto",
		YmlSettingsShutdownMessage:            "Server closed",
		YmlMemoryMainHardLimit:                1024,
		YmlMemoryGarbageCollectionPeriod:      36000,
		YmlMemoryMaxChunksChunkRadius:         4,
		YmlDebugLevel:                         1,
		YmlPlayerSavePlayerData:               true,
		"chunk-sending.per-tick":              4,
		"chunk-sending.spawn-radius":          4,
		YmlChunkGenerationPopulationQueueSize: 32,
		YmlTicksPerAutosave:                   6000,
		YmlTimingsHost:                        "timings.pmmp.io",
		YmlConsoleTitleTick:                   true,
		YmlPluginsLegacyDataDir:               false,
		YmlAutoUpdaterOnUpdateWarnConsole:     true,
		"network.max-mtu-size":                1492,
	} {
		if got := c(key); got != want {
			t.Errorf("%s = %v (%T), want %v", key, got, got, want)
		}
	}
	for _, table := range []string{YmlAliases, YmlWorlds} {
		if _, ok := c(table).(map[string]any); !ok {
			t.Errorf("%s isn't a table", table)
		}
	}
}
