package item

// GlassBottle is a port of pocketmine\item\GlassBottle. OnInteractBlock (filling with water to
// produce a Potion) needs a real Player/Block/World - see the Item interface's doc comment on
// Player/Entity-interaction methods.
type GlassBottle struct {
	ItemBase
}

func NewGlassBottle(identifier ItemIdentifier, name string) *GlassBottle {
	g := &GlassBottle{}
	g.Init(g, identifier, name)
	return g
}

func (g *GlassBottle) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}
