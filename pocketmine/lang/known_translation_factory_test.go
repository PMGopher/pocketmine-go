package lang

import "testing"

func TestVendoredLanguageTranslatesNamedAndPositionalParameters(t *testing.T) {
	l, err := NewLanguage("eng", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Translate(KnownTranslationFactory.PocketmineDisconnectError("Internal server error", "ab12-cd34")); got != "Internal server error (Error ID: ab12-cd34)" {
		t.Errorf("named parameters: got %q", got)
	}
	if got := l.Translate(KnownTranslationFactory.DeathAttackPlayer("Steve", "Alex")); got != "Steve was slain by Alex" {
		t.Errorf("positional parameters: got %q", got)
	}
	nested := KnownTranslationFactory.PocketmineDisconnectError(KnownTranslationFactory.PocketmineDisconnectErrorLoginTimeout(), "x").Prefix("§c")
	if got := l.Translate(nested); got != "§cLogin timeout (Error ID: x)" {
		t.Errorf("prefixed nested translation: got %q", got)
	}
	list, err := GetLanguageList("")
	if err != nil || list["eng"] != "English" {
		t.Errorf("GetLanguageList: %v %v", list["eng"], err)
	}
}
