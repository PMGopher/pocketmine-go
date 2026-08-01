package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestDyedShulkerBox(w World) *DyedShulkerBox {
	d := NewDyedShulkerBox(mustBlockIdentifier(1078), "Test Dyed Shulker Box", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	d.SetPosition(w, 1, 2, 3)
	return d
}

func TestDyedShulkerBoxDefaultColorIsWhite(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDyedShulkerBox(w)
	if d.GetColor() != blockutils.DyeColorWhite {
		t.Errorf("GetColor() = %v, want White", d.GetColor())
	}
}

func TestDyedShulkerBoxSetColor(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDyedShulkerBox(w)
	d.SetColor(blockutils.DyeColorRed)
	if d.GetColor() != blockutils.DyeColorRed {
		t.Errorf("GetColor() = %v, want Red", d.GetColor())
	}
}

// TestDyedShulkerBoxPlaceUsesShulkerBoxLogic confirms the embedded ShulkerBox.Place (faces the
// clicked face directly, not the player's opposite facing) is promoted correctly to
// DyedShulkerBox, which defines no Place override of its own.
func TestDyedShulkerBoxPlaceUsesShulkerBoxLogic(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDyedShulkerBox(w)
	tx := &fakeBlockTransaction{}

	d.Place(tx, fakeItem{}, d, d, math.East, math.Vector3{}, &fakeSignPlayer{})

	if d.Facing != math.East {
		t.Errorf("Facing = %v, want East (the clicked face)", d.Facing)
	}
}
