package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

// SmithingTableInventory is a port of pocketmine\block\inventory\SmithingTableInventory.
type SmithingTableInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewSmithingTableInventory(holder block.Position) *SmithingTableInventory {
	return &SmithingTableInventory{
		SimpleInventory:     inventory.NewSimpleInventory(3),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}
