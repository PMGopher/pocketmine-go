// Package blockinventory is a port of pocketmine\block\inventory.
package blockinventory

import "pocketmine-go/pocketmine/block"

// BlockInventory is a port of pocketmine\block\inventory\BlockInventory.
type BlockInventory interface {
	GetHolder() block.Position
}

// BlockInventoryTrait is a port of pocketmine\block\inventory\BlockInventoryTrait.
type BlockInventoryTrait struct {
	Holder block.Position
}

func (b *BlockInventoryTrait) GetHolder() block.Position { return b.Holder }
