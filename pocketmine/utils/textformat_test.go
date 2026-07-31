package utils

import "testing"

func TestTokenize(t *testing.T) {
	got := Tokenize(Red + "Hello " + Reset + "World")
	want := []string{Red, "Hello ", Reset, "World"}
	if len(got) != len(want) {
		t.Fatalf("Tokenize() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Tokenize()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCleanRemovesFormatCodes(t *testing.T) {
	got := Clean(Red+"Hello"+Reset, true)
	if got != "Hello" {
		t.Fatalf("Clean() = %q, want %q", got, "Hello")
	}
}

func TestColorize(t *testing.T) {
	got := Colorize("&cHello", "&")
	want := Red + "Hello"
	if got != want {
		t.Fatalf("Colorize() = %q, want %q", got, want)
	}
}

func TestToHTML(t *testing.T) {
	got := ToHTML(Red + "Hi" + Reset)
	want := `<span style="color:#F55">Hi</span>`
	if got != want {
		t.Fatalf("ToHTML() = %q, want %q", got, want)
	}
}
