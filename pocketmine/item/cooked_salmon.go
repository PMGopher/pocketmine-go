package item

// CookedSalmon is a port of pocketmine\item\CookedSalmon.
type CookedSalmon struct {
	Food
}

func NewCookedSalmon(identifier ItemIdentifier, name string) *CookedSalmon {
	c := &CookedSalmon{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedSalmon) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedSalmon) GetFoodRestore() int { return 6 }

func (c *CookedSalmon) GetSaturationRestore() float64 { return 9.6 }
