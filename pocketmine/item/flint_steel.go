package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// FlintSteel is a port of pocketmine\item\FlintSteel.
type FlintSteel struct {
	Tool
}

func NewFlintSteel(identifier ItemIdentifier, name string, enchantmentTags ...string) *FlintSteel {
	f := &FlintSteel{}
	f.Init(f, identifier, name)
	f.enchantmentTags = enchantmentTags
	return f
}

func (f *FlintSteel) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FlintSteel) GetMaxDurability() int { return 65 }

// OnInteractBlock is a port of FlintSteel::onInteractBlock: fire is lit on the clicked air block.
func (f *FlintSteel) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if blockReplace.GetTypeId() != block.AIR {
		return ItemUseResultNone
	}
	pos := blockReplace.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return ItemUseResultNone
	}
	_ = world.SetBlock(pos, block.VanillaBlock("fire"))
	world.AddSound(pos.Add(0.5, 0.5, 0.5), sound.FlintSteelSound{})

	f.ApplyDamage(1)

	return ItemUseResultSuccess
}
