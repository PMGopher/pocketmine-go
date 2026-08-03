package player

import "testing"

func TestNewPlayerInfoCleansFormattingCodesFromUsername(t *testing.T) {
	info := NewPlayerInfo("§cSteve", "uuid-1", nil, "en_US", nil)
	if got := info.GetUsername(); got != "Steve" {
		t.Errorf("GetUsername() = %q, want %q (formatting codes stripped)", got, "Steve")
	}
}

func TestNewPlayerInfoGettersReturnConstructorArguments(t *testing.T) {
	extra := map[string]any{"foo": "bar"}
	info := NewPlayerInfo("Steve", "uuid-1", nil, "en_US", extra)

	if info.GetUUID() != "uuid-1" {
		t.Errorf("GetUUID() = %q, want %q", info.GetUUID(), "uuid-1")
	}
	if info.GetLocale() != "en_US" {
		t.Errorf("GetLocale() = %q, want %q", info.GetLocale(), "en_US")
	}
	if info.GetExtraData()["foo"] != "bar" {
		t.Errorf("GetExtraData() = %v, want map with foo=bar", info.GetExtraData())
	}
}

func TestXboxLivePlayerInfoCarriesXuidAndWithoutXboxDataStripsIt(t *testing.T) {
	info := NewXboxLivePlayerInfo("xuid-1", "Steve", "uuid-1", nil, "en_US", nil)
	if info.GetXuid() != "xuid-1" {
		t.Errorf("GetXuid() = %q, want %q", info.GetXuid(), "xuid-1")
	}
	if info.GetUsername() != "Steve" {
		t.Errorf("GetUsername() = %q, want %q", info.GetUsername(), "Steve")
	}

	stripped := info.WithoutXboxData()
	if stripped.GetUsername() != "Steve" || stripped.GetUUID() != "uuid-1" {
		t.Errorf("WithoutXboxData() = %+v, want the same username/uuid without XUID", stripped)
	}
}
