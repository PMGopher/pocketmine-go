package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

const StonecutterSlotInput = 0

// StonecutterInventory is a port of pocketmine\block\inventory\StonecutterInventory.
type StonecutterInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewStonecutterInventory(holder block.Position) *StonecutterInventory {
	i := &StonecutterInventory{
		SimpleInventory:     inventory.NewSimpleInventory(1),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	i.Init(i)
	return i
}
