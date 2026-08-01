package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

var (
	_ inventory.Inventory = (*ChestInventory)(nil)
	_ inventory.Inventory = (*DoubleChestInventory)(nil)
	_ BlockInventory      = (*ChestInventory)(nil)
	_ BlockInventory      = (*DoubleChestInventory)(nil)
)

func newTestChestItem(count int) item.Item {
	a := item.NewApple(item.NewItemIdentifier(item.APPLE), "Apple")
	a.SetCount(count)
	return a
}

func TestChestInventorySize(t *testing.T) {
	w := &fakeWorld{}
	c := NewChestInventory(block.NewPosition(0, 0, 0, w))
	if c.GetSize() != 27 {
		t.Errorf("GetSize() = %d, want 27", c.GetSize())
	}
}

func TestDoubleChestInventorySizeIsSumOfBothHalves(t *testing.T) {
	w := &fakeWorld{}
	left := NewChestInventory(block.NewPosition(0, 0, 0, w))
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	if d.GetSize() != 54 {
		t.Errorf("GetSize() = %d, want 54", d.GetSize())
	}
}

func TestDoubleChestInventoryRoutesToCorrectHalf(t *testing.T) {
	w := &fakeWorld{}
	left := NewChestInventory(block.NewPosition(0, 0, 0, w))
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	d.SetItem(0, newTestChestItem(5))   // left slot 0
	d.SetItem(30, newTestChestItem(10)) // right slot 3 (30 - 27)

	if left.GetItem(0).GetCount() != 5 {
		t.Errorf("left slot 0 = %d, want 5", left.GetItem(0).GetCount())
	}
	if right.GetItem(3).GetCount() != 10 {
		t.Errorf("right slot 3 = %d, want 10", right.GetItem(3).GetCount())
	}
	if d.GetItem(0).GetCount() != 5 || d.GetItem(30).GetCount() != 10 {
		t.Errorf("d.GetItem(0)=%d d.GetItem(30)=%d, want 5/10", d.GetItem(0).GetCount(), d.GetItem(30).GetCount())
	}
}

func TestDoubleChestInventoryContainsAcrossBothHalves(t *testing.T) {
	w := &fakeWorld{}
	left := NewChestInventory(block.NewPosition(0, 0, 0, w))
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	d.SetItem(0, newTestChestItem(2))
	d.SetItem(28, newTestChestItem(3))

	if !d.Contains(newTestChestItem(5)) {
		t.Error("expected Contains to find 5 total matching items across both halves")
	}
}

func TestDoubleChestInventoryGetLeftAndRightSide(t *testing.T) {
	w := &fakeWorld{}
	left := NewChestInventory(block.NewPosition(0, 0, 0, w))
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	if d.GetLeftSide() != left || d.GetRightSide() != right {
		t.Error("expected GetLeftSide/GetRightSide to return the original halves")
	}
}

func TestDoubleChestInventoryHolderIsLeftsHolder(t *testing.T) {
	w := &fakeWorld{}
	leftHolder := block.NewPosition(5, 6, 7, w)
	left := NewChestInventory(leftHolder)
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	if d.GetHolder() != leftHolder {
		t.Errorf("GetHolder() = %v, want %v", d.GetHolder(), leftHolder)
	}
}

func TestDoubleChestInventoryOnOpenAnimatesOnFirstViewer(t *testing.T) {
	w := &fakeWorld{}
	left := NewChestInventory(block.NewPosition(0, 0, 0, w))
	right := NewChestInventory(block.NewPosition(1, 0, 0, w))
	d := NewDoubleChestInventory(left, right)

	d.OnOpen(&fakePlayer{"a"})
	if len(w.sounds) != 1 {
		t.Fatalf("len(sounds) = %d, want 1", len(w.sounds))
	}

	d.OnOpen(&fakePlayer{"b"})
	if len(w.sounds) != 1 {
		t.Errorf("len(sounds) after second viewer = %d, want still 1", len(w.sounds))
	}
}
