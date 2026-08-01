package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

const (
	EnchantSlotInput = 0
	EnchantSlotLapis = 1
)

// EnchantInventory is a port of pocketmine\block\inventory\EnchantInventory, minus the actual
// enchanting-option generation: OnSlotChange's override (regenerating options via
// EnchantingHelper.GenerateOptions and firing PlayerEnchantingOptionsRequestEvent whenever the
// input slot changes) needs the unported item/enchantment package (EnchantingHelper/
// EnchantingOption) and a real Player (to read their enchantment seed and sync options back over
// the network) - see the Item interface's doc comment on Player/Entity-interaction methods for
// the latter. GetOutput/GetOption are dropped for the same reason (they only ever return
// something derived from that unported machinery). GetInput/GetLapis are plain slot accessors and
// stay real.
type EnchantInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait
}

func NewEnchantInventory(holder block.Position) *EnchantInventory {
	return &EnchantInventory{
		SimpleInventory:     inventory.NewSimpleInventory(2),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
}

func (e *EnchantInventory) GetInput() item.Item { return e.GetItem(EnchantSlotInput) }

func (e *EnchantInventory) GetLapis() item.Item { return e.GetItem(EnchantSlotLapis) }
