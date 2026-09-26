package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/sound"
)

// BarrelInventory is a port of pocketmine\block\inventory\BarrelInventory.
type BarrelInventory struct {
	*AnimatedInventory
}

func NewBarrelInventory(holder block.Position) *BarrelInventory {
	b := &BarrelInventory{}
	b.AnimatedInventory = NewAnimatedInventory(27, holder, sound.BarrelOpenSound{}, sound.BarrelCloseSound{}, b.animateBlock)
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	b.Init(b)
	return b
}

// animateBlock is a port of BarrelInventory::animateBlock: the barrel block is opened/closed.
func (b *BarrelInventory) animateBlock(isOpen bool) {
	world, err := b.Holder.GetWorld()
	if err != nil {
		return
	}
	if barrel, ok := world.GetBlockAt(b.Holder.FloorX(), b.Holder.FloorY(), b.Holder.FloorZ()).(*block.Barrel); ok {
		barrel.SetOpen(isOpen)
		_ = world.SetBlock(b.Holder, barrel)
	}
}
