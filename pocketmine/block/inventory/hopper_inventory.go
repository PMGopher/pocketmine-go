package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

// HopperInventory is a port of pocketmine\block\inventory\HopperInventory.
type HopperInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

// NewHopperInventory is a port of HopperInventory::__construct (PHP's default size is 5).
func NewHopperInventory(holder block.Position, size int) *HopperInventory {
	h := &HopperInventory{
		SimpleInventory:     inventory.NewSimpleInventory(size),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	h.Init(h)
	return h
}
