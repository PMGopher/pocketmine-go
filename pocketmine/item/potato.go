package item

import "pocketmine-go/pocketmine/block"

// Potato is a port of pocketmine\item\Potato.
type Potato struct {
	Food
}

func NewPotato(identifier ItemIdentifier, name string) *Potato {
	p := &Potato{}
	p.Init(p, identifier, name)
	return p
}

func (p *Potato) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Potato) GetFoodRestore() int { return 1 }

func (p *Potato) GetSaturationRestore() float64 { return 0.6 }

// GetBlock is a port of Potato::getBlock.
func (x *Potato) GetBlock() block.Behavior { return block.VanillaBlock("potatoes") }
