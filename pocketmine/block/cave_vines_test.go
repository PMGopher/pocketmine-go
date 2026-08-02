package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestCaveVines(w World) *CaveVines {
	c := NewCaveVines(mustBlockIdentifier(1097), "Test Cave Vines", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCaveVinesOnInteractReturnsTrueWithoutPickingWhenBerried(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCaveVines(w)
	c.Berries = true

	if !c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if !c.Berries {
		t.Error("expected berries to remain set (picking isn't ported, so state shouldn't change)")
	}
}

func TestCaveVinesOnInteractIgnoresNonFertilizerWhenBerryless(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCaveVines(w)

	if c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false for a non-fertilizer item")
	}
}

func TestCaveVinesOnInteractFertilizesWithBoneMeal(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCaveVines(w)
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if !c.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	grown, ok := w.lastSetBlock.(*CaveVines)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *CaveVines, got %T", w.lastSetBlock)
	}
	if !grown.Berries {
		t.Error("expected the grown state to have berries")
	}
}

func TestCaveVinesOnRandomTickRecomputesHead(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCaveVines(w)
	c.Head = false          // stale: no cave vines below by default, so head should become true
	c.Age = CaveVinesMaxAge // disables the random growth-roll branch, so only the head-recompute
	// SetBlock call can happen - keeps this test deterministic instead of ~10% flaky.

	c.OnRandomTick()

	grown, ok := w.lastSetBlock.(*CaveVines)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *CaveVines, got %T", w.lastSetBlock)
	}
	if !grown.Head {
		t.Error("expected Head to be recomputed to true")
	}
}

func TestCaveVinesOnRandomTickDoesNothingWhenHeadUnchanged(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCaveVines(w)
	c.Head = true           // already matches (no cave vines below -> head=true), and age growth is
	c.Age = CaveVinesMaxAge // gated off by max age, so no SetBlock should happen at all

	c.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Errorf("expected no SetBlock call, got %T", w.lastSetBlock)
	}
}
