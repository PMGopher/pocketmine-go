package item

// Carrot is a port of pocketmine\item\Carrot. GetBlock (should return VanillaBlocks.CARROTS())
// isn't ported - GetBlock isn't part of the Item interface here at all yet (see the Item
// interface's doc comment), and needs the unported block registry regardless.
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
