package inventory

import "pocketmine-go/pocketmine/item"

// CallbackInventoryListener is a port of pocketmine\inventory\CallbackInventoryListener. Either
// callback may be nil.
type CallbackInventoryListener struct {
	onSlotChange    func(inv Inventory, slot int, oldItem item.Item)
	onContentChange func(inv Inventory, oldContents map[int]item.Item)
}

func NewCallbackInventoryListener(onSlotChange func(inv Inventory, slot int, oldItem item.Item), onContentChange func(inv Inventory, oldContents map[int]item.Item)) *CallbackInventoryListener {
	return &CallbackInventoryListener{onSlotChange: onSlotChange, onContentChange: onContentChange}
}

// OnAnyChange is a port of CallbackInventoryListener::onAnyChange: onChange runs for both slot and
// whole-content changes.
func OnAnyChange(onChange func(inv Inventory)) *CallbackInventoryListener {
	return NewCallbackInventoryListener(
		func(inv Inventory, _ int, _ item.Item) { onChange(inv) },
		func(inv Inventory, _ map[int]item.Item) { onChange(inv) },
	)
}

func (c *CallbackInventoryListener) OnSlotChange(inv Inventory, slot int, oldItem item.Item) {
	if c.onSlotChange != nil {
		c.onSlotChange(inv, slot, oldItem)
	}
}

func (c *CallbackInventoryListener) OnContentChange(inv Inventory, oldContents map[int]item.Item) {
	if c.onContentChange != nil {
		c.onContentChange(inv, oldContents)
	}
}
