package lang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLangFile(t *testing.T, dir, code, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, code+".ini"), []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write test ini file: %v", err)
	}
}

func TestLanguageBasicGetAndFallback(t *testing.T) {
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", "language.name = English\ngreeting = Hello, {%0}!\n")
	writeLangFile(t, dir, "fra", "language.name = French\n") // missing "greeting" -> falls back to eng

	l, err := NewLanguage("fra", dir, "eng")
	if err != nil {
		t.Fatalf("NewLanguage() error = %v", err)
	}
	if l.Name() != "French" {
		t.Fatalf("Name() = %q, want French", l.Name())
	}
	got, _ := l.TranslateString("greeting", []any{"Steve"}, nil)
	if got != "Hello, Steve!" {
		t.Fatalf("TranslateString() = %q, want %q", got, "Hello, Steve!")
	}
}

func TestLanguageUnknownKeyReturnsKeyItself(t *testing.T) {
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", "language.name = English\n")

	l, err := NewLanguage("eng", dir, "eng")
	if err != nil {
		t.Fatalf("NewLanguage() error = %v", err)
	}
	if got := l.Get("no.such.key"); got != "no.such.key" {
		t.Fatalf("Get(missing) = %q, want the key back unchanged", got)
	}
}

func TestLanguageEmbeddedKeySubstitution(t *testing.T) {
	// Embedded "%key" substitution (parseTranslation) only kicks in when the text passed to
	// TranslateString/Translate is NOT itself a registered translation key — e.g. a raw composed
	// string like a command usage hint, not a plain "get me this key" lookup. Both PHP and this
	// port skip it entirely when the input string is found directly as a key.
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", strings.TrimSpace(`
language.name = English
color.red = RED
`)+"\n")

	l, err := NewLanguage("eng", dir, "eng")
	if err != nil {
		t.Fatalf("NewLanguage() error = %v", err)
	}
	got, _ := l.TranslateString("%color.red status: {%0}", []any{"ok"}, nil)
	if got != "RED status: ok" {
		t.Fatalf("TranslateString() = %q, want %q", got, "RED status: ok")
	}
}

func TestLanguageTranslateNestedTranslatable(t *testing.T) {
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", "language.name = English\ngreeting = Hello, {%0}!\nplayer.notch = Notch\n")

	l, err := NewLanguage("eng", dir, "eng")
	if err != nil {
		t.Fatalf("NewLanguage() error = %v", err)
	}

	inner := NewTranslatable("player.notch", nil)
	outer := NewTranslatable("greeting", []any{inner})
	got := l.Translate(outer)
	if got != "Hello, Notch!" {
		t.Fatalf("Translate() = %q, want %q", got, "Hello, Notch!")
	}
}

func TestGetLanguageList(t *testing.T) {
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", "language.name = English\n")
	writeLangFile(t, dir, "deu", "language.name = Deutsch\n")

	list, err := GetLanguageList(dir)
	if err != nil {
		t.Fatalf("GetLanguageList() error = %v", err)
	}
	if list["eng"] != "English" || list["deu"] != "Deutsch" {
		t.Fatalf("GetLanguageList() = %v", list)
	}
}

func TestNewLanguageMissingFileReturnsLanguageNotFoundException(t *testing.T) {
	dir := t.TempDir()
	writeLangFile(t, dir, "eng", "language.name = English\n")

	_, err := NewLanguage("xyz", dir, "eng")
	if err == nil {
		t.Fatalf("expected an error for a missing language file")
	}
	if _, ok := err.(*LanguageNotFoundException); !ok {
		t.Fatalf("expected *LanguageNotFoundException, got %T", err)
	}
}
