package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/world/sound"
)

// ShulkerBoxInventory is a port of pocketmine\block\inventory\ShulkerBoxInventory.
type ShulkerBoxInventory struct {
	*AnimatedInventory
}

func NewShulkerBoxInventory(holder block.Position) *ShulkerBoxInventory {
	s := &ShulkerBoxInventory{}
	s.AnimatedInventory = NewAnimatedInventory(27, holder, sound.ShulkerBoxOpenSound{}, sound.ShulkerBoxCloseSound{}, s.animateBlock)
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	s.Init(s)
	return s
}

// CanAddItem is a port of ShulkerBoxInventory::canAddItem: shulker boxes can't go in a shulker box.
func (s *ShulkerBoxInventory) CanAddItem(it item.Item) bool {
	// ItemTypeIds::toBlockTypeId: block items have negative type IDs.
	if typeID := it.GetTypeId(); typeID <= 0 {
		if blockTypeID := -typeID; blockTypeID == block.SHULKER_BOX || blockTypeID == block.DYED_SHULKER_BOX {
			return false
		}
	}
	return s.AnimatedInventory.CanAddItem(it)
}

// animateBlock is a port of ShulkerBoxInventory::animateBlock.
func (s *ShulkerBoxInventory) animateBlock(isOpen bool) { broadcastChestEvent(s.Holder, isOpen) }
