package player

import (
	"testing"

	"pocketmine-go/pocketmine/nbt"
)

func TestOfflinePlayerWithNilNamedTagHasNoData(t *testing.T) {
	p := NewOfflinePlayer("Steve", nil)
	if p.HasPlayedBefore() {
		t.Error("HasPlayedBefore() = true with a nil namedTag")
	}
	if _, ok := p.GetFirstPlayed(); ok {
		t.Error("GetFirstPlayed() ok = true with a nil namedTag")
	}
	if _, ok := p.GetLastPlayed(); ok {
		t.Error("GetLastPlayed() ok = true with a nil namedTag")
	}
}

func TestOfflinePlayerReadsFirstAndLastPlayedFromNBT(t *testing.T) {
	tag := nbt.NewCompoundTag().SetLong(TagFirstPlayed, 1000).SetLong(TagLastPlayed, 2000)
	p := NewOfflinePlayer("Steve", tag)

	if !p.HasPlayedBefore() {
		t.Fatal("HasPlayedBefore() = false with a non-nil namedTag")
	}
	first, ok := p.GetFirstPlayed()
	if !ok || first != 1000 {
		t.Errorf("GetFirstPlayed() = (%d, %v), want (1000, true)", first, ok)
	}
	last, ok := p.GetLastPlayed()
	if !ok || last != 2000 {
		t.Errorf("GetLastPlayed() = (%d, %v), want (2000, true)", last, ok)
	}
}

func TestOfflinePlayerGetNameReturnsTheConstructorName(t *testing.T) {
	p := NewOfflinePlayer("Notch", nil)
	if p.GetName() != "Notch" {
		t.Errorf("GetName() = %q, want %q", p.GetName(), "Notch")
	}
}
