package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

func newTestTrappedChest(w World, x, y, z int) *TrappedChest {
	c := NewTrappedChest(mustBlockIdentifier(1072), "Test Trapped Chest", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, x, y, z)
	return c
}

func placeTrappedChestWithTile(w *chestWorld, x, y, z int, facing math.Facing) *TrappedChest {
	c := newTestTrappedChest(w, x, y, z)
	c.Facing = facing
	t := tile.NewChest(w, math.NewVector3(float64(x), float64(y), float64(z)))
	w.tiles[[3]int{x, y, z}] = t
	w.blockTiles[[3]int{x, y, z}] = t
	w.blocks[[3]int{x, y, z}] = c
	return c
}

// TestTrappedChestOnPostPlacePairsWithAdjacentTrappedChest exercises the exact bug this file's
// doc comment describes: OnPostPlace's neighbor-pairing type assertion has to target
// *TrappedChest, not *Chest, or two adjacent trapped chests would silently fail to pair despite
// PHP's instanceof-based check succeeding for this same case.
func TestTrappedChestOnPostPlacePairsWithAdjacentTrappedChest(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	placeTrappedChestWithTile(w, 1, 2, 3, math.North)
	c2 := placeTrappedChestWithTile(w, 2, 2, 3, math.North)

	c2.OnPostPlace()

	t1, _ := w.GetTile(NewPosition(1, 2, 3, w))
	t2, _ := w.GetTile(NewPosition(2, 2, 3, w))
	tc1, tc2 := t1.(*tile.Chest), t2.(*tile.Chest)
	if !tc1.IsPaired() || !tc2.IsPaired() {
		t.Fatal("expected both trapped chest tiles to be paired after OnPostPlace")
	}
}

// TestTrappedChestOnPostPlaceDoesNotPairWithPlainChest mirrors hasSameTypeId's real effect in the
// PHP original: a TrappedChest and a plain Chest never pair, even though TrappedChest instanceof
// Chest is true, because their block type IDs differ.
func TestTrappedChestOnPostPlaceDoesNotPairWithPlainChest(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	placeChestWithTile(w, 1, 2, 3, math.North)
	c2 := placeTrappedChestWithTile(w, 2, 2, 3, math.North)

	c2.OnPostPlace()

	t1, _ := w.GetTile(NewPosition(1, 2, 3, w))
	t2, _ := w.GetTile(NewPosition(2, 2, 3, w))
	tc1, tc2 := t1.(*tile.Chest), t2.(*tile.Chest)
	if tc1.IsPaired() || tc2.IsPaired() {
		t.Error("expected a TrappedChest and a plain Chest not to pair")
	}
}
