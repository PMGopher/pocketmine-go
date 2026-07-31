package utils

import (
	"path/filepath"
	"testing"
)

func TestConfigYAMLRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pocketmine.yml")

	cfg, err := NewConfig(file, ConfigDetect, map[string]any{
		"settings": map[string]any{"motd": "Test server", "max-players": int64(20)},
	})
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if got := cfg.GetNested("settings.motd", nil); got != "Test server" {
		t.Fatalf("GetNested(settings.motd) = %v, want %q", got, "Test server")
	}

	cfg.SetNested("settings.max-players", int64(50))
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reloaded, err := NewConfig(file, ConfigDetect, nil)
	if err != nil {
		t.Fatalf("reload NewConfig() error = %v", err)
	}
	if got := reloaded.GetNested("settings.max-players", nil); got != 50 {
		t.Fatalf("after reload, GetNested(settings.max-players) = %v (%T), want 50", got, got)
	}
}

func TestConfigPropertiesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "server.properties")

	cfg, err := NewConfig(file, ConfigDetect, map[string]any{
		"server-port": int64(19132),
		"gamemode":    "survival",
	})
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	cfg.Set("enable-query", true)
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reloaded, err := NewConfig(file, ConfigDetect, nil)
	if err != nil {
		t.Fatalf("reload error = %v", err)
	}
	if got := reloaded.Get("gamemode", nil); got != "survival" {
		t.Fatalf("Get(gamemode) = %v, want survival", got)
	}
	if got := reloaded.Get("enable-query", nil); got != true {
		t.Fatalf("Get(enable-query) = %v, want true", got)
	}
	if got := reloaded.Get("server-port", nil); got != int64(19132) {
		t.Fatalf("Get(server-port) = %v (%T), want int64(19132)", got, got)
	}
}

func TestConfigEnumRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "whitelist.txt")

	cfg, err := NewConfig(file, ConfigDetect, nil)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	cfg.Set("Notch", true)
	cfg.Set("Steve", true)
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reloaded, err := NewConfig(file, ConfigDetect, nil)
	if err != nil {
		t.Fatalf("reload error = %v", err)
	}
	if !reloaded.Exists("Notch", false) || !reloaded.Exists("Steve", false) {
		t.Fatalf("expected both entries to round-trip, got keys %v", reloaded.GetKeys())
	}
}
