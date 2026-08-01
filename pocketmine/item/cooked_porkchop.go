package item

// CookedPorkchop is a port of pocketmine\item\CookedPorkchop.
type CookedPorkchop struct {
	Food
}

func NewCookedPorkchop(identifier ItemIdentifier, name string) *CookedPorkchop {
	c := &CookedPorkchop{}
	c.Init(c, identifier, name)
	return c
}

func (c *CookedPorkchop) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CookedPorkchop) GetFoodRestore() int { return 8 }

func (c *CookedPorkchop) GetSaturationRestore() float64 { return 12.8 }
