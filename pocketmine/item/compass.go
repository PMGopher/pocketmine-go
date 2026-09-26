package item

// Compass is a port of pocketmine\item\Compass - an empty class body in the PHP original too.
type Compass struct {
	ItemBase
}

func NewCompass(identifier ItemIdentifier, name string, enchantmentTags ...string) *Compass {
	c := &Compass{}
	c.Init(c, identifier, name)
	c.enchantmentTags = enchantmentTags
	return c
}

func (c *Compass) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}
