package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/crafting"
)

// CraftingTableInventory is a port of pocketmine\block\inventory\CraftingTableInventory.
type CraftingTableInventory struct {
	*crafting.CraftingGrid
	BlockInventoryTrait
}

func NewCraftingTableInventory(holder block.Position) *CraftingTableInventory {
	return &CraftingTableInventory{
		CraftingGrid:        crafting.NewCraftingGrid(crafting.CraftingGridSizeBig),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}
