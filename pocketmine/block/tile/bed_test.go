package tile

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestBedSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	b := NewBed(w, math.Vector3{})
	b.SetColor(blockutils.DyeColorBlue)

	saved := b.SaveNBT()

	decoded := NewBed(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetColor() != blockutils.DyeColorBlue {
		t.Errorf("GetColor() = %v, want Blue", decoded.GetColor())
	}
}

func TestBedReadSaveDataDefaultsToRedOnMissingColor(t *testing.T) {
	w := &fakeWorld{}
	b := NewBed(w, math.Vector3{})
	b.SetColor(blockutils.DyeColorBlue) // should be overwritten by ReadSaveData's default

	if err := b.ReadSaveData(nbt.NewCompoundTag()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.GetColor() != blockutils.DyeColorRed {
		t.Errorf("GetColor() = %v, want Red (legacy pre-1.1 default)", b.GetColor())
	}
}

func TestBedReadSaveDataIgnoresOutOfRangeColor(t *testing.T) {
	w := &fakeWorld{}
	b := NewBed(w, math.Vector3{})
	b.SetColor(blockutils.DyeColorBlue)

	tag := nbt.NewCompoundTag()
	tag.SetByte(BedTagColor, nbt.ByteTag(99))

	if err := b.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.GetColor() != blockutils.DyeColorRed {
		t.Errorf("GetColor() = %v, want Red (out-of-range byte falls back to default)", b.GetColor())
	}
}
