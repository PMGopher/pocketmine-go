package tile

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestChiseledBookshelfGetSetLastInteractedSlot(t *testing.T) {
	w := &fakeWorld{}
	c := NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))

	if _, has := c.GetLastInteractedSlot(); has {
		t.Error("expected a fresh tile to have no last-interacted slot")
	}

	slot := blockutils.ChiseledBookshelfSlotTopRight
	c.SetLastInteractedSlot(&slot)

	got, has := c.GetLastInteractedSlot()
	if !has || got != blockutils.ChiseledBookshelfSlotTopRight {
		t.Errorf("GetLastInteractedSlot() = (%v, %v), want (TopRight, true)", got, has)
	}
}

func TestChiseledBookshelfSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	c := NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))
	slot := blockutils.ChiseledBookshelfSlotBottomLeft
	c.SetLastInteractedSlot(&slot)

	tag := nbt.NewCompoundTag()
	c.WriteSaveData(tag)

	loaded := NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))
	if err := loaded.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData() error = %v", err)
	}

	got, has := loaded.GetLastInteractedSlot()
	if !has || got != blockutils.ChiseledBookshelfSlotBottomLeft {
		t.Errorf("round-tripped slot = (%v, %v), want (BottomLeft, true)", got, has)
	}
}

func TestChiseledBookshelfSaveDataRoundTripWithNoSlot(t *testing.T) {
	w := &fakeWorld{}
	c := NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))

	tag := nbt.NewCompoundTag()
	c.WriteSaveData(tag)

	loaded := NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))
	if err := loaded.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData() error = %v", err)
	}

	if _, has := loaded.GetLastInteractedSlot(); has {
		t.Error("expected no last-interacted slot to round-trip as none")
	}
}
