package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// FireCharge is a port of pocketmine\item\FireCharge.
type FireCharge struct {
	ItemBase
}

func NewFireCharge(identifier ItemIdentifier, name string) *FireCharge {
	f := &FireCharge{}
	f.Init(f, identifier, name)
	return f
}

func (f *FireCharge) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

// OnInteractBlock is a port of FireCharge::onInteractBlock: fire is lit on the clicked air block.
func (f *FireCharge) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if blockReplace.GetTypeId() == block.AIR {
		pos := blockReplace.GetPosition()
		world, err := pos.GetWorld()
		if err != nil {
			return ItemUseResultNone
		}
		_ = world.SetBlock(pos, block.VanillaBlock("fire"))
		world.AddSound(pos.Add(0.5, 0.5, 0.5), sound.BlazeShootSound{})

		f.Pop()

		return ItemUseResultSuccess
	}
	return ItemUseResultNone
}
