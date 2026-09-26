package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

// CartographyTableInventory is a port of pocketmine\block\inventory\CartographyTableInventory.
type CartographyTableInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewCartographyTableInventory(holder block.Position) *CartographyTableInventory {
	i := &CartographyTableInventory{
		SimpleInventory:     inventory.NewSimpleInventory(2),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	i.Init(i)
	return i
}
