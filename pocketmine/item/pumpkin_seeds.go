package item

import "pocketmine-go/pocketmine/block"

// PumpkinSeeds is a port of pocketmine\item\PumpkinSeeds.
type PumpkinSeeds struct {
	ItemBase
}

func NewPumpkinSeeds(identifier ItemIdentifier, name string) *PumpkinSeeds {
	p := &PumpkinSeeds{}
	p.Init(p, identifier, name)
	return p
}

func (p *PumpkinSeeds) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

// GetBlock is a port of PumpkinSeeds::getBlock.
func (x *PumpkinSeeds) GetBlock() block.Behavior { return block.VanillaBlock("pumpkin_stem") }
