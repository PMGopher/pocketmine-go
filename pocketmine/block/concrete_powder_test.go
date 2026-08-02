package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

func newTestConcretePowder(w World) *ConcretePowder {
	c := NewConcretePowder(mustBlockIdentifier(1095), "Test Concrete Powder", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestConcretePowderOnNearbyBlockChangeFormsConcreteWithSameColor(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestConcretePowder(w)
	c.SetColor(blockutils.DyeColorRed)

	water := newTestWater(w)
	water.SetPosition(w, 1, 2, 2) // North of the powder (North offset is (0,0,-1))
	w.blocks[[3]int{1, 2, 2}] = water

	c.OnNearbyBlockChange()

	concrete, ok := w.lastSetBlock.(*Concrete)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *Concrete, got %T", w.lastSetBlock)
	}
	if concrete.GetColor() != blockutils.DyeColorRed {
		t.Errorf("GetColor() = %v, want Red", concrete.GetColor())
	}
}

func TestConcretePowderOnNearbyBlockChangeDoesNothingWithoutWater(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestConcretePowder(w)

	c.OnNearbyBlockChange()

	if w.lastSetBlock != nil {
		t.Errorf("expected no block change, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestConcretePowderTickFallingReturnsConcreteWhenAdjacentToWater(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestConcretePowder(w)
	c.SetColor(blockutils.DyeColorBlue)

	water := newTestWater(w)
	water.SetPosition(w, 2, 2, 3) // East of the powder
	w.blocks[[3]int{2, 2, 3}] = water

	replacement, ok := c.TickFalling()
	if !ok {
		t.Fatal("expected TickFalling to report a replacement")
	}
	concrete, ok := replacement.(*Concrete)
	if !ok {
		t.Fatalf("replacement = %T, want *Concrete", replacement)
	}
	if concrete.GetColor() != blockutils.DyeColorBlue {
		t.Errorf("GetColor() = %v, want Blue", concrete.GetColor())
	}
}

func TestConcretePowderTickFallingKeepsFallingWithoutWater(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestConcretePowder(w)

	if _, ok := c.TickFalling(); ok {
		t.Error("expected TickFalling to report no replacement without adjacent water")
	}
}
