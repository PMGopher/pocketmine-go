package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

// CampfireInventory is a port of pocketmine\block\inventory\CampfireInventory - a plain 4-slot
// SimpleInventory with a BlockInventoryTrait holder and a max stack size of 1, same no-animation
// shape as LoomInventory.
type CampfireInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewCampfireInventory(holder block.Position) *CampfireInventory {
	c := &CampfireInventory{
		SimpleInventory:     inventory.NewSimpleInventory(4),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	c.SetMaxStackSize(1)
	return c
}
