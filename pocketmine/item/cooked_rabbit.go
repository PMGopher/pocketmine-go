package item

// CookedRabbit is a port of pocketmine\item\CookedRabbit.
type CookedRabbit struct {
	Food
}

func NewCookedRabbit(identifier ItemIdentifier, name string) *CookedRabbit {
	c := &CookedRabbit{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedRabbit) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedRabbit) GetFoodRestore() int { return 5 }

func (c *CookedRabbit) GetSaturationRestore() float64 { return 6 }
