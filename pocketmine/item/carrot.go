package item

import "pocketmine-go/pocketmine/block"

// Carrot is a port of pocketmine\item\Carrot.
type Carrot struct {
	Food
}

func NewCarrot(identifier ItemIdentifier, name string) *Carrot {
	c := &Carrot{}
	c.Init(c, identifier, name)
	return c
}

func (c *Carrot) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Carrot) GetFoodRestore() int { return 3 }

func (c *Carrot) GetSaturationRestore() float64 { return 4.8 }

// GetBlock is a port of Carrot::getBlock.
func (x *Carrot) GetBlock() block.Behavior { return block.VanillaBlock("carrots") }
