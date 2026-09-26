package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

const (
	BrewingStandSlotIngredient   = 0
	BrewingStandSlotBottleLeft   = 1
	BrewingStandSlotBottleMiddle = 2
	BrewingStandSlotBottleRight  = 3
	BrewingStandSlotFuel         = 4
)

// BrewingStandInventory is a port of pocketmine\block\inventory\BrewingStandInventory.
type BrewingStandInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

// NewBrewingStandInventory is a port of BrewingStandInventory::__construct (PHP's default size
// is 5).
func NewBrewingStandInventory(holder block.Position, size int) *BrewingStandInventory {
	b := &BrewingStandInventory{
		SimpleInventory:     inventory.NewSimpleInventory(size),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	b.Init(b)
	return b
}
