package item

// CookedFish is a port of pocketmine\item\CookedFish.
type CookedFish struct {
	Food
}

func NewCookedFish(identifier ItemIdentifier, name string) *CookedFish {
	c := &CookedFish{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedFish) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedFish) GetFoodRestore() int { return 5 }

func (c *CookedFish) GetSaturationRestore() float64 { return 6 }
