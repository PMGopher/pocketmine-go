package inventory

import "pocketmine-go/pocketmine/item"

// DelegateInventory is a port of pocketmine\inventory\DelegateInventory: an inventory which is
// backed by another inventory, and acts as a proxy to that inventory. Changes to the backing
// inventory are reported to this inventory's own listeners and viewers.
type DelegateInventory struct {
	BaseInventory

	backingInventory         Inventory
	inventoryListener        InventoryListener
	backingInventoryChanging bool
}

func NewDelegateInventory(backingInventory Inventory) *DelegateInventory {
	d := &DelegateInventory{backingInventory: backingInventory}
	d.Init(d)
	d.inventoryListener = NewCallbackInventoryListener(
		func(_ Inventory, slot int, oldItem item.Item) {
			d.backingInventoryChanging = true
			defer func() { d.backingInventoryChanging = false }()
			d.BaseInventory.onSlotChange(slot, oldItem)
		},
		func(_ Inventory, oldContents map[int]item.Item) {
			d.backingInventoryChanging = true
			defer func() { d.backingInventoryChanging = false }()
			d.BaseInventory.onContentChange(oldContents)
		},
	)
	backingInventory.GetListeners().Add(d.inventoryListener)
	return d
}

// Close detaches this inventory from its backing inventory (PHP's __destruct).
func (d *DelegateInventory) Close() {
	d.backingInventory.GetListeners().Remove(d.inventoryListener)
}

func (d *DelegateInventory) GetSize() int { return d.backingInventory.GetSize() }

func (d *DelegateInventory) GetItem(index int) item.Item { return d.backingInventory.GetItem(index) }

// InternalSetItem forwards to the backing inventory, whose listener reports the change.
func (d *DelegateInventory) InternalSetItem(index int, it item.Item) {
	d.backingInventory.SetItem(index, it)
}

func (d *DelegateInventory) GetContents(includeEmpty bool) map[int]item.Item {
	return d.backingInventory.GetContents(includeEmpty)
}

func (d *DelegateInventory) InternalSetContents(items map[int]item.Item) {
	d.backingInventory.SetContents(items)
}

func (d *DelegateInventory) IsSlotEmpty(index int) bool { return d.backingInventory.IsSlotEmpty(index) }

// reportsChanges is DelegateInventory's onSlotChange/onContentChange override: changes are only
// reported when they come from the backing inventory's listener (which every change does, since
// this inventory writes through to it).
func (d *DelegateInventory) reportsChanges() bool { return d.backingInventoryChanging }
