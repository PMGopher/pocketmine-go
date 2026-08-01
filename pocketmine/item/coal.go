package item

// Coal is a port of pocketmine\item\Coal.
type Coal struct {
	ItemBase
}

func NewCoal(identifier ItemIdentifier, name string) *Coal {
	c := &Coal{}
	c.Init(c, identifier, name)
	return c
}

func (c *Coal) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Coal) GetFuelTime() int { return 1600 }
