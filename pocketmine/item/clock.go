package item

// Clock is a port of pocketmine\item\Clock - an empty class body in the PHP original too.
type Clock struct {
	ItemBase
}

func NewClock(identifier ItemIdentifier, name string) *Clock {
	c := &Clock{}
	c.Init(c, identifier, name)
	return c
}

func (c *Clock) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}
