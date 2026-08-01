package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/world/sound"
)

// AnimatedInventory is a port of pocketmine\block\inventory\AnimatedBlockInventoryTrait, combined
// with SimpleInventory and BlockInventoryTrait the way a concrete container inventory (e.g.
// ChestInventory) composes them in PHP.
//
// The PHP trait's three abstract hooks (getOpenSound/getCloseSound/animateBlock) are per-instance
// configuration in practice - each concrete container just returns a fixed sound pair and calls a
// specific tile method - so they're constructor-provided values/a callback here instead of a
// separate self-dispatch layer, avoiding an interface with no second implementation to justify it.
type AnimatedInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait

	OpenSound    sound.Sound
	CloseSound   sound.Sound
	AnimateBlock func(isOpen bool)
}

func NewAnimatedInventory(size int, holder block.Position, openSound, closeSound sound.Sound, animateBlock func(isOpen bool)) *AnimatedInventory {
	return &AnimatedInventory{
		SimpleInventory:     inventory.NewSimpleInventory(size),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
		OpenSound:           openSound,
		CloseSound:          closeSound,
		AnimateBlock:        animateBlock,
	}
}

func (a *AnimatedInventory) GetViewerCount() int { return len(a.GetViewers()) }

// OnOpen is a port of AnimatedBlockInventoryTrait::onOpen.
func (a *AnimatedInventory) OnOpen(who inventory.Player) {
	a.SimpleInventory.OnOpen(who)

	if a.Holder.IsValid() && a.GetViewerCount() == 1 {
		// TODO: this crap really shouldn't be managed by the inventory (comment preserved from PHP)
		if a.AnimateBlock != nil {
			a.AnimateBlock(true)
		}
		if world, err := a.Holder.GetWorld(); err == nil {
			world.AddSound(a.Holder.Add(0.5, 0.5, 0.5), a.OpenSound)
		}
	}
}

// OnClose is a port of AnimatedBlockInventoryTrait::onClose.
func (a *AnimatedInventory) OnClose(who inventory.Player) {
	if a.Holder.IsValid() && a.GetViewerCount() == 1 {
		if a.AnimateBlock != nil {
			a.AnimateBlock(false)
		}
		if world, err := a.Holder.GetWorld(); err == nil {
			world.AddSound(a.Holder.Add(0.5, 0.5, 0.5), a.CloseSound)
		}
	}

	a.SimpleInventory.OnClose(who)
}
