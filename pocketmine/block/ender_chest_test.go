package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

func newTestEnderChest(w World) *EnderChest {
	e := NewEnderChest(mustBlockIdentifier(1068), "Test Ender Chest", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	e.SetPosition(w, 1, 2, 3)
	return e
}

func TestEnderChestGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	e := newTestEnderChest(w)
	if e.GetLightLevel() != 7 {
		t.Errorf("GetLightLevel() = %d, want 7", e.GetLightLevel())
	}
}

func TestEnderChestIsAffectedBySilkTouch(t *testing.T) {
	w := &fakeWorld{}
	e := newTestEnderChest(w)
	if !e.IsAffectedBySilkTouch() {
		t.Error("expected EnderChest to be affected by silk touch")
	}
}

func TestEnderChestOnInteractIncrementsViewerCountWhenUnblocked(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	e := newTestEnderChest(w)
	tileEnder := tile.NewEnderChest(w, math.NewVector3(1, 2, 3))
	w.tiles[[3]int{1, 2, 3}] = tileEnder

	if !e.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if tileEnder.GetViewerCount() != 1 {
		t.Errorf("GetViewerCount() = %d, want 1", tileEnder.GetViewerCount())
	}
}

func TestEnderChestOnInteractDoesNotIncrementWhenLidBlocked(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	e := newTestEnderChest(w)
	tileEnder := tile.NewEnderChest(w, math.NewVector3(1, 2, 3))
	w.tiles[[3]int{1, 2, 3}] = tileEnder

	opaque := newTestBlock(false)
	opaque.SetPosition(w, 1, 3, 3)
	w.blocks = map[[3]int]Behavior{{1, 3, 3}: opaque}

	if !e.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if tileEnder.GetViewerCount() != 0 {
		t.Errorf("GetViewerCount() = %d, want 0 (blocked lid)", tileEnder.GetViewerCount())
	}
}
