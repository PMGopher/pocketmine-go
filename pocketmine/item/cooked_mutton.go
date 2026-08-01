package item

// CookedMutton is a port of pocketmine\item\CookedMutton.
type CookedMutton struct {
	Food
}

func NewCookedMutton(identifier ItemIdentifier, name string) *CookedMutton {
	c := &CookedMutton{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedMutton) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedMutton) GetFoodRestore() int { return 6 }

func (c *CookedMutton) GetSaturationRestore() float64 { return 9.6 }
