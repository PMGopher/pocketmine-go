package item

import "pocketmine-go/pocketmine/block"

// Redstone is a port of pocketmine\item\Redstone.
type Redstone struct {
	ItemBase
}

func NewRedstone(identifier ItemIdentifier, name string) *Redstone {
	r := &Redstone{}
	r.Init(r, identifier, name)
	return r
}

func (r *Redstone) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

// GetBlock is a port of Redstone::getBlock.
func (x *Redstone) GetBlock() block.Behavior { return block.VanillaBlock("redstone_wire") }
