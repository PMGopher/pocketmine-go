package item

// CookedChicken is a port of pocketmine\item\CookedChicken.
type CookedChicken struct {
	Food
}

func NewCookedChicken(identifier ItemIdentifier, name string) *CookedChicken {
	c := &CookedChicken{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedChicken) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedChicken) GetFoodRestore() int { return 6 }

func (c *CookedChicken) GetSaturationRestore() float64 { return 7.2 }
