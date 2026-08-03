package chat

import (
	"testing"

	"pocketmine-go/pocketmine/lang"
)

func TestStandardChatFormatterReturnsATranslatableWithUsernameAndMessage(t *testing.T) {
	got := StandardChatFormatter{}.Format("Steve", "hello")
	tr, ok := got.(*lang.Translatable)
	if !ok {
		t.Fatalf("Format() returned %T, want *lang.Translatable", got)
	}
	if tr.Text() != chatTypeTextKey {
		t.Errorf("Text() = %q, want %q", tr.Text(), chatTypeTextKey)
	}
	if tr.Parameter(0) != "Steve" || tr.Parameter(1) != "hello" {
		t.Errorf("Parameters() = %v, want [Steve hello]", tr.Parameters())
	}
}

func TestLegacyRawChatFormatterSubstitutesPlaceholders(t *testing.T) {
	f := NewLegacyRawChatFormatter("<{%0}> {%1}")
	got := f.Format("Steve", "hello")
	if got != "<Steve> hello" {
		t.Errorf("Format() = %q, want %q", got, "<Steve> hello")
	}
}

func TestLegacyRawChatFormatterWithRepeatedPlaceholders(t *testing.T) {
	f := NewLegacyRawChatFormatter("{%0}: {%1} (from {%0})")
	got := f.Format("Steve", "hi")
	if got != "Steve: hi (from Steve)" {
		t.Errorf("Format() = %q, want %q", got, "Steve: hi (from Steve)")
	}
}
