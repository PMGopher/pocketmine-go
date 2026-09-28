package item

import "pocketmine-go/pocketmine/block"

// BeetrootSeeds is a port of pocketmine\item\BeetrootSeeds.
type BeetrootSeeds struct {
	ItemBase
}

func NewBeetrootSeeds(identifier ItemIdentifier, name string) *BeetrootSeeds {
	b := &BeetrootSeeds{}
	b.Init(b, identifier, name)
	return b
}

func (b *BeetrootSeeds) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

// GetBlock is a port of BeetrootSeeds::getBlock.
func (x *BeetrootSeeds) GetBlock() block.Behavior { return block.VanillaBlock("beetroots") }
