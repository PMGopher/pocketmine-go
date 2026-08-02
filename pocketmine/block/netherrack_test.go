package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestNetherrack(w World) *Netherrack {
	n := NewNetherrack(mustBlockIdentifier(1098), "Test Netherrack", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	n.SetPosition(w, 1, 2, 3)
	return n
}

func placeTypedBlockAt(w *containerTileWorld, x, y, z, typeID int) {
	b := newFakeTypedBlock(typeID)
	b.SetPosition(w, x, y, z)
	w.blocks[[3]int{x, y, z}] = b
}

func TestNetherrackTryTransformFailsWhenCoveredAbove(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)
	solid := newTestBlock(false) // IsTransparent() == false
	solid.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = solid

	if n.tryTransform() {
		t.Error("expected tryTransform to fail when covered above by a non-transparent block")
	}
}

func TestNetherrackTryTransformFailsWithoutNearbyNylium(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)

	if n.tryTransform() {
		t.Error("expected tryTransform to fail with no nearby nylium")
	}
}

func TestNetherrackTryTransformBecomesWarpedNyliumWhenOnlyWarpedNearby(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)
	placeTypedBlockAt(w, 2, 2, 3, WARPED_NYLIUM)

	if !n.tryTransform() {
		t.Fatal("expected tryTransform to succeed")
	}
	if _, ok := w.lastSetBlock.(*Nylium); !ok {
		t.Fatalf("expected SetBlock to be called with a *Nylium, got %T", w.lastSetBlock)
	}
	if w.lastSetBlock.GetTypeId() != WARPED_NYLIUM {
		t.Errorf("GetTypeId() = %d, want WARPED_NYLIUM (%d)", w.lastSetBlock.GetTypeId(), WARPED_NYLIUM)
	}
}

func TestNetherrackTryTransformBecomesCrimsonNyliumWhenOnlyCrimsonNearby(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)
	placeTypedBlockAt(w, 0, 2, 3, CRIMSON_NYLIUM)

	if !n.tryTransform() {
		t.Fatal("expected tryTransform to succeed")
	}
	if w.lastSetBlock.GetTypeId() != CRIMSON_NYLIUM {
		t.Errorf("GetTypeId() = %d, want CRIMSON_NYLIUM (%d)", w.lastSetBlock.GetTypeId(), CRIMSON_NYLIUM)
	}
}

func TestNetherrackOnInteractFertilizesWithBoneMeal(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)
	placeTypedBlockAt(w, 2, 2, 3, WARPED_NYLIUM)
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if !n.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
}

func TestNetherrackOnInteractIgnoresNonFertilizerItems(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	n := newTestNetherrack(w)
	placeTypedBlockAt(w, 2, 2, 3, WARPED_NYLIUM)

	if n.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false for a non-fertilizer item")
	}
}
