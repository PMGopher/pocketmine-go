package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

const (
	AnvilSlotInput    = 0
	AnvilSlotMaterial = 1
)

// AnvilInventory is a port of pocketmine\block\inventory\AnvilInventory.
type AnvilInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewAnvilInventory(holder block.Position) *AnvilInventory {
	a := &AnvilInventory{
		SimpleInventory:     inventory.NewSimpleInventory(2),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	a.Init(a)
	return a
}
