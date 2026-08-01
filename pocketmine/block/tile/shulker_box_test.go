package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestShulkerBoxDefaultFacingAndName(t *testing.T) {
	w := &fakeWorld{}
	s := NewShulkerBox(w, math.Vector3{})
	if s.GetFacing() != int(math.North) {
		t.Errorf("GetFacing() = %d, want North", s.GetFacing())
	}
	if s.GetName() != "Shulker Box" {
		t.Errorf("GetName() = %q, want %q", s.GetName(), "Shulker Box")
	}
}

func TestShulkerBoxSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	s := NewShulkerBox(w, math.Vector3{})
	s.SetFacing(int(math.Up))
	s.SetName("Loot")

	decoded := NewShulkerBox(w, math.Vector3{})
	if err := decoded.ReadSaveData(s.SaveNBT()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetFacing() != int(math.Up) {
		t.Errorf("GetFacing() = %d, want Up", decoded.GetFacing())
	}
	if decoded.GetName() != "Loot" {
		t.Errorf("GetName() = %q, want %q", decoded.GetName(), "Loot")
	}
}

func TestShulkerBoxCopyDataFromItemUsesWholeItemNBT(t *testing.T) {
	w := &fakeWorld{}
	source := NewShulkerBox(w, math.Vector3{})
	source.SetFacing(int(math.South))

	item := fakeItem{blockNbt: source.SaveNBT(), hasBlkNbt: true, customName: "Renamed", hasName: true}

	dest := NewShulkerBox(w, math.Vector3{})
	dest.CopyDataFromItem(item)

	if dest.GetFacing() != int(math.South) {
		t.Errorf("GetFacing() = %d, want South", dest.GetFacing())
	}
	if dest.GetName() != "Renamed" {
		t.Errorf("GetName() = %q, want %q", dest.GetName(), "Renamed")
	}
}
