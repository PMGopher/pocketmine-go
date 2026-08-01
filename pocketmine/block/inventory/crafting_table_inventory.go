package blockinventory

import "pocketmine-go/pocketmine/block"

// CraftingTableInventory is a port of pocketmine\block\inventory\CraftingTableInventory.
type CraftingTableInventory struct {
	*CraftingGrid
	BlockInventoryTrait
}

func NewCraftingTableInventory(holder block.Position) *CraftingTableInventory {
	return &CraftingTableInventory{
		CraftingGrid:        NewCraftingGrid(CraftingGridSizeBig),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}
