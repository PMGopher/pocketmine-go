package server

import (
	"path/filepath"
	"testing"

	"pocketmine-go/pocketmine/utils"
)

func newTestConfigGroup(t *testing.T, args []string) *ServerConfigGroup {
	t.Helper()
	props, err := utils.NewConfig(filepath.Join(t.TempDir(), "server.properties"), utils.ConfigProperties, map[string]any{
		PropertyMotd:           "Test Server",
		PropertyServerPortIPv4: 19132,
		PropertyXboxAuth:       "on",
		PropertyPvp:            false,
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewServerConfigGroup(nil, props, args)
}

func TestGetConfigValuesFromServerProperties(t *testing.T) {
	g := newTestConfigGroup(t, nil)
	if got := g.GetConfigString(PropertyMotd, ""); got != "Test Server" {
		t.Errorf("motd = %q", got)
	}
	if got := g.GetConfigInt(PropertyServerPortIPv4, 0); got != 19132 {
		t.Errorf("server-port = %d", got)
	}
	if !g.GetConfigBool(PropertyXboxAuth, false) {
		t.Error(`xbox-auth "on" should be true`)
	}
	if g.GetConfigBool(PropertyPvp, true) {
		t.Error("pvp false should be false")
	}
	if got := g.GetConfigInt(PropertyMaxPlayers, 20); got != 20 {
		t.Errorf("missing max-players = %d, want the default 20", got)
	}
}

func TestCommandLineOverridesServerProperties(t *testing.T) {
	// ServerConfigGroup's getopt("", ["$variable::"]).
	g := newTestConfigGroup(t, []string{"--server-port=19133", "--xbox-auth=false", "--motd=Other"})
	if got := g.GetConfigInt(PropertyServerPortIPv4, 0); got != 19133 {
		t.Errorf("server-port = %d, want the --server-port override", got)
	}
	if g.GetConfigBool(PropertyXboxAuth, true) {
		t.Error("--xbox-auth=false should override the file")
	}
	if got := g.GetConfigString(PropertyMotd, ""); got != "Other" {
		t.Errorf("motd = %q, want the --motd override", got)
	}
}

func TestSetConfigAndSave(t *testing.T) {
	g := newTestConfigGroup(t, nil)
	g.SetConfigBool(PropertyWhitelist, true)
	g.SetConfigInt(PropertyMaxPlayers, 5)
	if !g.GetConfigBool(PropertyWhitelist, false) || g.GetConfigInt(PropertyMaxPlayers, 0) != 5 {
		t.Error("set values aren't read back")
	}
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyDefaultsWithoutPocketmineYml(t *testing.T) {
	g := newTestConfigGroup(t, []string{"--chunk-sending.per-tick=8"})
	if got := g.GetPropertyInt("chunk-sending.per-tick", 4); got != 8 {
		t.Errorf("per-tick = %d, want the command-line value 8", got)
	}
	if got := g.GetPropertyInt("chunk-sending.spawn-radius", 4); got != 4 {
		t.Errorf("spawn-radius = %d, want the default 4", got)
	}
}
