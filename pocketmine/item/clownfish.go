package item

// Clownfish is a port of pocketmine\item\Clownfish.
type Clownfish struct {
	Food
}

func NewClownfish(identifier ItemIdentifier, name string) *Clownfish {
	c := &Clownfish{}
	c.Init(c, identifier, name)
	return c
}

func (c *Clownfish) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Clownfish) GetFoodRestore() int { return 1 }

func (c *Clownfish) GetSaturationRestore() float64 { return 0.2 }
