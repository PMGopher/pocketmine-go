package inventory

import "pocketmine-go/pocketmine/item"

// PlayerCursorInventory is a port of pocketmine\inventory\PlayerCursorInventory: the item held by
// the mouse cursor while an inventory is open.
type PlayerCursorInventory struct {
	SimpleInventory

	holder Player
}

func NewPlayerCursorInventory(holder Player) *PlayerCursorInventory {
	c := &PlayerCursorInventory{SimpleInventory: SimpleInventory{slots: make([]item.Item, 1)}, holder: holder}
	c.Init(c)
	return c
}

func (c *PlayerCursorInventory) GetHolder() Player { return c.holder }

func (c *PlayerCursorInventory) IsTemporaryInventory() {}
