package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/sound"
)

// ChestInventory is a port of pocketmine\block\inventory\ChestInventory.
//
// animateBlock needs network\mcpe\protocol\BlockEventPacket (not ported), so it's a documented
// no-op callback rather than actually broadcasting anything - the sound and viewer-count
// bookkeeping in AnimatedInventory's OnOpen/OnClose still work correctly regardless.
type ChestInventory struct {
	*AnimatedInventory
}

func NewChestInventory(holder block.Position) *ChestInventory {
	c := &ChestInventory{}
	c.AnimatedInventory = NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, c.animateBlock)
	return c
}

// animateBlock should broadcast a BlockEventPacket (event ID 1, always, for a chest) to the
// holder's viewers - see the doc comment above for why this is a no-op.
func (c *ChestInventory) animateBlock(isOpen bool) {}
