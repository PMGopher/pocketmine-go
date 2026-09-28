package item

import "pocketmine-go/pocketmine/block"

// TorchflowerSeeds is a port of pocketmine\item\TorchflowerSeeds.
type TorchflowerSeeds struct {
	ItemBase
}

func NewTorchflowerSeeds(identifier ItemIdentifier, name string) *TorchflowerSeeds {
	t := &TorchflowerSeeds{}
	t.Init(t, identifier, name)
	return t
}

func (t *TorchflowerSeeds) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}

// GetBlock is a port of TorchflowerSeeds::getBlock.
func (x *TorchflowerSeeds) GetBlock() block.Behavior { return block.VanillaBlock("torchflower_crop") }
