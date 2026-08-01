package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

func newTestBrewingStand(w World) *BrewingStand {
	b := NewBrewingStand(mustBlockIdentifier(1071), "Test Brewing Stand", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBrewingStandSlotGetSlotNumber(t *testing.T) {
	cases := []struct {
		slot BrewingStandSlot
		want int
	}{
		{BrewingStandSlotEast, 1},
		{BrewingStandSlotNorthwest, 2},
		{BrewingStandSlotSouthwest, 3},
	}
	for _, c := range cases {
		if got := c.slot.GetSlotNumber(); got != c.want {
			t.Errorf("%v.GetSlotNumber() = %d, want %d", c.slot, got, c.want)
		}
	}
}

func TestBrewingStandSetSlotAndHasSlot(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBrewingStand(w)

	if b.HasSlot(BrewingStandSlotEast) {
		t.Error("expected a fresh brewing stand to have no occupied slots")
	}
	b.SetSlot(BrewingStandSlotEast, true)
	if !b.HasSlot(BrewingStandSlotEast) {
		t.Error("expected SetSlot(East, true) to mark it occupied")
	}
	b.SetSlot(BrewingStandSlotEast, false)
	if b.HasSlot(BrewingStandSlotEast) {
		t.Error("expected SetSlot(East, false) to clear it")
	}
}

func TestBrewingStandCloneDeepCopiesSlots(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBrewingStand(w)
	b.SetSlot(BrewingStandSlotNorthwest, true)

	clone := b.Clone().(*BrewingStand)
	clone.SetSlot(BrewingStandSlotSouthwest, true)

	if b.HasSlot(BrewingStandSlotSouthwest) {
		t.Error("expected cloning not to leak mutations back to the original")
	}
	if !clone.HasSlot(BrewingStandSlotNorthwest) {
		t.Error("expected the clone to retain the original's occupied slots")
	}
}

func TestBrewingStandOnInteractCompletesWithTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	b := newTestBrewingStand(w)
	w.tiles[[3]int{1, 2, 3}] = tile.NewBrewingStand(w, math.NewVector3(1, 2, 3))

	if !b.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}
