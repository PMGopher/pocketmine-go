package inventory

import (
	"testing"

	"pocketmine-go/pocketmine/item"
)

var _ Inventory = (*SimpleInventory)(nil)

func newTestItem(count int) item.Item {
	a := item.NewApple(item.NewItemIdentifier(item.APPLE), "Apple")
	a.SetCount(count)
	return a
}

func TestSimpleInventoryGetSetItemRoundTrip(t *testing.T) {
	inv := NewSimpleInventory(9)
	inv.SetItem(3, newTestItem(5))

	got := inv.GetItem(3)
	if got.IsNull() || got.GetCount() != 5 {
		t.Errorf("GetItem(3) count = %d, want 5", got.GetCount())
	}
}

func TestSimpleInventoryGetItemReturnsIndependentClone(t *testing.T) {
	inv := NewSimpleInventory(9)
	inv.SetItem(0, newTestItem(5))

	got := inv.GetItem(0)
	got.SetCount(1)

	if inv.GetItem(0).GetCount() != 5 {
		t.Error("expected mutating the item returned by GetItem not to affect the inventory")
	}
}

func TestSimpleInventoryEmptySlotReturnsNullItem(t *testing.T) {
	inv := NewSimpleInventory(9)
	if !inv.GetItem(0).IsNull() {
		t.Error("expected an empty slot to return a null item")
	}
	if !inv.IsSlotEmpty(0) {
		t.Error("expected IsSlotEmpty(0) to be true for a fresh inventory")
	}
}

func TestSimpleInventorySetItemWithNullClearsSlot(t *testing.T) {
	inv := NewSimpleInventory(9)
	inv.SetItem(0, newTestItem(5))
	inv.SetItem(0, newTestItem(0))

	if !inv.IsSlotEmpty(0) {
		t.Error("expected setting a null item to clear the slot")
	}
}

func TestSimpleInventoryAddItemFillsExistingStackThenEmptySlots(t *testing.T) {
	inv := NewSimpleInventory(2)
	inv.SetItem(0, newTestItem(60)) // Apple max stack size is 64 (ItemBase default)

	leftover := inv.AddItem(newTestItem(10))
	if len(leftover) != 0 {
		t.Fatalf("expected no leftover, got %v", leftover)
	}
	if got := inv.GetItem(0).GetCount(); got != 64 {
		t.Errorf("slot 0 count = %d, want 64 (topped up)", got)
	}
	if got := inv.GetItem(1).GetCount(); got != 6 {
		t.Errorf("slot 1 count = %d, want 6 (overflow into empty slot)", got)
	}
}

func TestSimpleInventoryAddItemReturnsLeftoverWhenFull(t *testing.T) {
	inv := NewSimpleInventory(1)
	inv.SetItem(0, newTestItem(64))

	leftover := inv.AddItem(newTestItem(5))
	if len(leftover) != 1 || leftover[0].GetCount() != 5 {
		t.Fatalf("leftover = %v, want a single 5-count item", leftover)
	}
}

func TestSimpleInventoryRemoveItem(t *testing.T) {
	inv := NewSimpleInventory(2)
	inv.SetItem(0, newTestItem(10))
	inv.SetItem(1, newTestItem(10))

	unremoved := inv.RemoveItem(newTestItem(15))
	if len(unremoved) != 0 {
		t.Fatalf("expected everything to be removed, got leftover %v", unremoved)
	}
	if inv.GetItem(0).GetCount()+inv.GetItem(1).GetCount() != 5 {
		t.Errorf("expected 5 items remaining across both slots, got %d + %d", inv.GetItem(0).GetCount(), inv.GetItem(1).GetCount())
	}
}

func TestSimpleInventoryRemoveItemReturnsUnremovedWhenNotEnough(t *testing.T) {
	inv := NewSimpleInventory(1)
	inv.SetItem(0, newTestItem(3))

	unremoved := inv.RemoveItem(newTestItem(10))
	if len(unremoved) != 1 || unremoved[0].GetCount() != 7 {
		t.Fatalf("unremoved = %v, want a single 7-count remainder", unremoved)
	}
	if !inv.IsSlotEmpty(0) {
		t.Error("expected the slot to be emptied of what could be removed")
	}
}

func TestSimpleInventoryContainsAndAll(t *testing.T) {
	inv := NewSimpleInventory(3)
	inv.SetItem(0, newTestItem(2))
	inv.SetItem(2, newTestItem(3))

	if !inv.Contains(newTestItem(4)) {
		t.Error("expected Contains to find 4 total matching items across slots 0 and 2")
	}
	if inv.Contains(newTestItem(6)) {
		t.Error("expected Contains to fail when requesting more than available")
	}

	all := inv.All(newTestItem(1))
	if len(all) != 2 {
		t.Errorf("len(All()) = %d, want 2", len(all))
	}
}

func TestSimpleInventoryFirstAndFirstEmpty(t *testing.T) {
	inv := NewSimpleInventory(3)
	inv.SetItem(1, newTestItem(5))

	if idx := inv.First(newTestItem(1), false); idx != 1 {
		t.Errorf("First() = %d, want 1", idx)
	}
	if idx := inv.First(newTestItem(10), false); idx != -1 {
		t.Errorf("First() for an unsatisfiable count = %d, want -1", idx)
	}
	if idx := inv.FirstEmpty(); idx != 0 {
		t.Errorf("FirstEmpty() = %d, want 0", idx)
	}
}

func TestSimpleInventoryCanAddItemAndGetAddableItemQuantity(t *testing.T) {
	inv := NewSimpleInventory(1)
	inv.SetItem(0, newTestItem(60))

	if inv.CanAddItem(newTestItem(10)) {
		t.Error("expected CanAddItem to fail: only 4 more fit in slot 0 and there are no empty slots")
	}
	if got := inv.GetAddableItemQuantity(newTestItem(10)); got != 4 {
		t.Errorf("GetAddableItemQuantity() = %d, want 4", got)
	}
}

func TestSimpleInventorySwap(t *testing.T) {
	inv := NewSimpleInventory(2)
	inv.SetItem(0, newTestItem(1))
	inv.SetItem(1, newTestItem(2))

	inv.Swap(0, 1)

	if inv.GetItem(0).GetCount() != 2 || inv.GetItem(1).GetCount() != 1 {
		t.Errorf("after swap: slot0=%d slot1=%d, want 2/1", inv.GetItem(0).GetCount(), inv.GetItem(1).GetCount())
	}
}

func TestSimpleInventoryClearAndClearAll(t *testing.T) {
	inv := NewSimpleInventory(2)
	inv.SetItem(0, newTestItem(1))
	inv.SetItem(1, newTestItem(1))

	inv.Clear(0)
	if !inv.IsSlotEmpty(0) || inv.IsSlotEmpty(1) {
		t.Fatal("expected only slot 0 to be cleared")
	}

	inv.ClearAll()
	if !inv.IsSlotEmpty(1) {
		t.Error("expected ClearAll to empty every slot")
	}
}

func TestSimpleInventorySlotExists(t *testing.T) {
	inv := NewSimpleInventory(3)
	if !inv.SlotExists(0) || !inv.SlotExists(2) {
		t.Error("expected slots 0 and 2 to exist")
	}
	if inv.SlotExists(-1) || inv.SlotExists(3) {
		t.Error("expected out-of-range slots not to exist")
	}
}

type recordingListener struct {
	slotChanges    int
	contentChanges int
}

func (r *recordingListener) OnSlotChange(inv Inventory, index int, before item.Item) {
	r.slotChanges++
}

func (r *recordingListener) OnContentChange(inv Inventory, before map[int]item.Item) {
	r.contentChanges++
}

func TestSimpleInventoryListenersFire(t *testing.T) {
	inv := NewSimpleInventory(2)
	l := &recordingListener{}
	inv.GetListeners().Add(l)

	inv.SetItem(0, newTestItem(1))
	if l.slotChanges != 1 {
		t.Errorf("slotChanges = %d, want 1", l.slotChanges)
	}

	inv.SetContents(map[int]item.Item{0: newTestItem(2)})
	if l.contentChanges != 1 {
		t.Errorf("contentChanges = %d, want 1", l.contentChanges)
	}
}

type fakePlayer struct{ name string }

func TestSimpleInventoryViewers(t *testing.T) {
	inv := NewSimpleInventory(1)
	p1 := &fakePlayer{"a"}
	p2 := &fakePlayer{"b"}

	inv.OnOpen(p1)
	inv.OnOpen(p2)
	if len(inv.GetViewers()) != 2 {
		t.Fatalf("len(GetViewers()) = %d, want 2", len(inv.GetViewers()))
	}

	inv.OnClose(p1)
	viewers := inv.GetViewers()
	if len(viewers) != 1 || viewers[0] != Player(p2) {
		t.Errorf("GetViewers() after close = %v, want [p2]", viewers)
	}
}
