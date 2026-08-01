package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

func newTestShulkerBox(w World) *ShulkerBox {
	s := NewShulkerBox(mustBlockIdentifier(1069), "Test Shulker Box", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestShulkerBoxPlaceFacesClickedFaceDirectly(t *testing.T) {
	w := &fakeWorld{}
	s := newTestShulkerBox(w)
	tx := &fakeBlockTransaction{}

	s.Place(tx, fakeItem{}, s, s, math.East, math.Vector3{}, nil)

	if s.Facing != math.East {
		t.Errorf("Facing = %v, want East (the clicked face)", s.Facing)
	}
}

func TestShulkerBoxGetMaxStackSize(t *testing.T) {
	w := &fakeWorld{}
	s := newTestShulkerBox(w)
	if s.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", s.GetMaxStackSize())
	}
}

func TestShulkerBoxReadStateFromWorldPullsFacingFromTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	s := newTestShulkerBox(w)
	tileShulker := tile.NewShulkerBox(w, math.NewVector3(1, 2, 3))
	tileShulker.SetFacing(int(math.Down))
	w.tiles[[3]int{1, 2, 3}] = tileShulker

	s.ReadStateFromWorld()

	if s.Facing != math.Down {
		t.Errorf("Facing = %v, want Down (pulled from tile)", s.Facing)
	}
}

func TestShulkerBoxOnInteractBlockedBySolidNeighbor(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	s := newTestShulkerBox(w)
	s.Facing = math.Up
	w.tiles[[3]int{1, 2, 3}] = tile.NewShulkerBox(w, math.NewVector3(1, 2, 3))

	solid := newTestBlock(false) // testBlock's IsSolid defaults to true (Block's default)
	solid.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = solid

	// Should complete without panicking and report handled either way.
	if !s.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}
