package item

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

var (
	_ Item = (*Boat)(nil)
	_ Item = (*Record)(nil)
	_ Item = (*Minecart)(nil)
)

func TestBoatTypeMatchesWoodType(t *testing.T) {
	if BoatTypeMangrove.GetWoodType() != blockutils.WoodTypeMangrove {
		t.Errorf("BoatTypeMangrove.GetWoodType() = %v, want WoodTypeMangrove", BoatTypeMangrove.GetWoodType())
	}
	if got := BoatTypeOak.GetDisplayName(); got != blockutils.WoodTypeOak.GetDisplayName() {
		t.Errorf("BoatTypeOak.GetDisplayName() = %q, want %q", got, blockutils.WoodTypeOak.GetDisplayName())
	}
}

func TestBoatProperties(t *testing.T) {
	b := NewBoat(NewItemIdentifier(OAK_BOAT), "Oak Boat", BoatTypeOak)
	if b.GetType() != BoatTypeOak {
		t.Errorf("GetType() = %v, want Oak", b.GetType())
	}
	if b.GetFuelTime() != 1200 {
		t.Errorf("GetFuelTime() = %d, want 1200", b.GetFuelTime())
	}
	if b.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", b.GetMaxStackSize())
	}
}

func TestRecordProperties(t *testing.T) {
	r := NewRecord(NewItemIdentifier(RECORD_13), blockutils.RecordTypeDisk13, "Music Disc 13")
	if r.GetRecordType() != blockutils.RecordTypeDisk13 {
		t.Errorf("GetRecordType() = %v, want Disk13", r.GetRecordType())
	}
	if r.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", r.GetMaxStackSize())
	}
	if got := blockutils.RecordTypeDisk13.GetSoundName(); got != "C418 - 13" {
		t.Errorf("GetSoundName() = %q, want %q", got, "C418 - 13")
	}
	if blockutils.RecordTypeDisk13.GetSoundId() != 0 {
		t.Error("expected the deprecated GetSoundId() to always return 0")
	}
}

func TestMinecartMaxStackSize(t *testing.T) {
	m := NewMinecart(NewItemIdentifier(MINECART), "Minecart")
	if m.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", m.GetMaxStackSize())
	}
}
