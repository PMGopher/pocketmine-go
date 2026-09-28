package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/plugin"
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
		"plugin.yml":           {Data: []byte("name: ExamplePlugin\nversion: 1.2.3\nmain: example\\Main\napi: [5.0.0]\nload: POSTWORLD\nauthors: [Alex]\n")},
		"resources/config.yml": {Data: []byte("enabled: true\n")},
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
	// pocketmine.yml has plugins.legacy-data-dir: false, so the data folder is in plugin_data.
	if want := filepath.Join(s.GetDataPath(), "plugin_data", "ExamplePlugin"); !strings.HasPrefix(p.GetDataFolder(), want) {
		t.Errorf("data folder = %q, want %q", p.GetDataFolder(), want)
	}
	if _, err := os.Stat(filepath.Join(p.GetDataFolder(), "config.yml")); err != nil {
		t.Errorf("the embedded config.yml wasn't saved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.GetDataPath(), "plugin_list.yml")); err != nil {
		t.Errorf("plugin_list.yml wasn't created: %v", err)
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
	if err := os.WriteFile(filepath.Join(dir, "plugin_list.yml"), []byte("mode: blacklist\nplugins: [ExamplePlugin]\n"), 0o644); err != nil {
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
