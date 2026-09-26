package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/sound"
)

// ChestInventory is a port of pocketmine\block\inventory\ChestInventory.
type ChestInventory struct {
	*AnimatedInventory
}

func NewChestInventory(holder block.Position) *ChestInventory {
	c := &ChestInventory{}
	c.AnimatedInventory = NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, c.animateBlock)
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	c.Init(c)
	return c
}

// animateBlock is a port of ChestInventory::animateBlock.
func (c *ChestInventory) animateBlock(isOpen bool) { broadcastChestEvent(c.Holder, isOpen) }
