package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestChiseledBookshelf(w World) *ChiseledBookshelf {
	c := NewChiseledBookshelf(mustBlockIdentifier(1085), "Test Chiseled Bookshelf", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestChiseledBookshelfHasSlotAndSetSlot(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)

	if c.HasSlot(blockutils.ChiseledBookshelfSlotTopLeft) {
		t.Error("expected a fresh bookshelf to have no occupied slots")
	}
	c.SetSlot(blockutils.ChiseledBookshelfSlotTopLeft, true)
	if !c.HasSlot(blockutils.ChiseledBookshelfSlotTopLeft) {
		t.Error("expected SetSlot(TopLeft, true) to mark it occupied")
	}
	c.SetSlot(blockutils.ChiseledBookshelfSlotTopLeft, false)
	if c.HasSlot(blockutils.ChiseledBookshelfSlotTopLeft) {
		t.Error("expected SetSlot(TopLeft, false) to clear it")
	}
}

func TestChiseledBookshelfCloneDeepCopiesSlotsAndLastInteracted(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)
	c.SetSlot(blockutils.ChiseledBookshelfSlotBottomRight, true)
	slot := blockutils.ChiseledBookshelfSlotBottomRight
	c.LastInteractedSlot = &slot

	clone := c.Clone().(*ChiseledBookshelf)
	clone.SetSlot(blockutils.ChiseledBookshelfSlotTopLeft, true)
	*clone.LastInteractedSlot = blockutils.ChiseledBookshelfSlotTopLeft

	if c.HasSlot(blockutils.ChiseledBookshelfSlotTopLeft) {
		t.Error("expected cloning not to leak slot mutations back to the original")
	}
	if *c.LastInteractedSlot != blockutils.ChiseledBookshelfSlotBottomRight {
		t.Error("expected cloning not to leak LastInteractedSlot mutations back to the original")
	}
}

func TestChiseledBookshelfPlaceFacesOppositePlayer(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)
	tx := &fakeBlockTransaction{}
	player := &fakeSignPlayer{}

	c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, player)

	if c.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player facing", c.Facing)
	}
}

func TestChiseledBookshelfOnInteractRejectsWrongFace(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)
	c.Facing = math.North

	if c.OnInteract(fakeItem{}, math.South, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false when the clicked face isn't the block's facing")
	}
}

func TestChiseledBookshelfOnInteractAcceptsMatchingFace(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)
	c.Facing = math.North

	if !c.OnInteract(fakeItem{}, math.North, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return true when the clicked face matches the block's facing")
	}
}

func TestChiseledBookshelfReadStateFromWorldPullsLastInteractedSlotFromTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestChiseledBookshelf(w)

	tileShelf := tile.NewChiseledBookshelf(w, math.NewVector3(1, 2, 3))
	slot := blockutils.ChiseledBookshelfSlotBottomMiddle
	tileShelf.SetLastInteractedSlot(&slot)
	w.tiles[[3]int{1, 2, 3}] = tileShelf

	c.ReadStateFromWorld()

	if c.LastInteractedSlot == nil || *c.LastInteractedSlot != blockutils.ChiseledBookshelfSlotBottomMiddle {
		t.Errorf("LastInteractedSlot = %v, want BottomMiddle", c.LastInteractedSlot)
	}
}

func TestChiseledBookshelfGetDropsForCompatibleToolIsEmpty(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChiseledBookshelf(w)
	if drops := c.GetDropsForCompatibleTool(fakeItem{}); len(drops) != 0 {
		t.Errorf("GetDropsForCompatibleTool() = %v, want empty", drops)
	}
}
