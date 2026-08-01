package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

const (
	LoomSlotBanner  = 0
	LoomSlotDye     = 1
	LoomSlotPattern = 2
)

// LoomInventory is a port of pocketmine\block\inventory\LoomInventory - a plain 3-slot
// SimpleInventory with a BlockInventoryTrait holder, unlike ChestInventory/DoubleChestInventory
// there's no open/close sound or animation for a loom, so this just embeds
// inventory.SimpleInventory directly.
type LoomInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewLoomInventory(holder block.Position) *LoomInventory {
	return &LoomInventory{
		SimpleInventory:     inventory.NewSimpleInventory(3),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}
