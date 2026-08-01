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
	return &CartographyTableInventory{
		SimpleInventory:     inventory.NewSimpleInventory(2),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}
