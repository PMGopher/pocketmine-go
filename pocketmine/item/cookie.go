package item

// Cookie is a port of pocketmine\item\Cookie.
type Cookie struct {
	Food
}

func NewCookie(identifier ItemIdentifier, name string) *Cookie {
	c := &Cookie{}
	c.Init(c, identifier, name)
	return c
}

func (c *Cookie) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Cookie) GetFoodRestore() int { return 2 }

func (c *Cookie) GetSaturationRestore() float64 { return 0.4 }
