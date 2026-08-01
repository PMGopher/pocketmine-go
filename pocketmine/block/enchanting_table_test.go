package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestEnchantingTable(w World) *EnchantingTable {
	e := NewEnchantingTable(mustBlockIdentifier(1075), "Test Enchanting Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	e.SetPosition(w, 1, 2, 3)
	return e
}

func TestEnchantingTableGetSupportType(t *testing.T) {
	w := &fakeWorld{}
	e := newTestEnchantingTable(w)
	if got := e.GetSupportType(math.Down); got != blockutils.SupportTypeNone {
		t.Errorf("GetSupportType(Down) = %v, want None", got)
	}
}

func TestEnchantingTableRecalculateCollisionBoxesTrimsTop(t *testing.T) {
	w := &fakeWorld{}
	e := newTestEnchantingTable(w)
	boxes := e.RecalculateCollisionBoxes()
	if len(boxes) != 1 {
		t.Fatalf("len(boxes) = %d, want 1", len(boxes))
	}
	want := math.OneAABB().TrimmedCopy(math.Up, 0.25)
	if boxes[0] != want {
		t.Errorf("boxes[0] = %+v, want %+v", boxes[0], want)
	}
}

func TestEnchantingTableOnInteractReturnsTrue(t *testing.T) {
	w := &fakeWorld{}
	e := newTestEnchantingTable(w)
	if !e.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}
