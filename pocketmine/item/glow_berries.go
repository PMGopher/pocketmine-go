package item

// GlowBerries is a port of pocketmine\item\GlowBerries.
type GlowBerries struct {
	Food
}

func NewGlowBerries(identifier ItemIdentifier, name string) *GlowBerries {
	g := &GlowBerries{}
	g.Init(g, identifier, name)
	return g
}

func (g *GlowBerries) Clone() Item {
	cl := *g
	cl.rebind(&cl)
	return &cl
}

func (g *GlowBerries) GetFoodRestore() int { return 2 }

func (g *GlowBerries) GetSaturationRestore() float64 { return 0.4 }
