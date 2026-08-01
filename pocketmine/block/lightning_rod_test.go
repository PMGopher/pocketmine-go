package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestLightningRod(w World) *LightningRod {
	idInfo, err := NewBlockIdentifier(1006, nil)
	if err != nil {
		panic(err)
	}
	l := NewLightningRod(idInfo, "Test Lightning Rod", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	l.SetPosition(w, 1, 2, 3)
	return l
}

func TestLightningRodPlaceSetsFacing(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLightningRod(w)
	tx := &fakeBlockTransaction{}

	if !l.Place(tx, fakeItem{}, l, l, math.East, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed")
	}
	if l.Facing != math.East {
		t.Errorf("Facing = %v, want East", l.Facing)
	}
}

func TestLightningRodHoneycombWaxesIt(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLightningRod(w)

	honeycomb := fakeItem{typeID: itemTypeIDsHoneycomb}
	if !l.OnInteract(honeycomb, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to apply wax")
	}
	if !l.IsWaxed() {
		t.Error("honeycomb should have waxed the lightning rod")
	}
}
