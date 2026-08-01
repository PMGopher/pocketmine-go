package item

// GoldenCarrot is a port of pocketmine\item\GoldenCarrot.
type GoldenCarrot struct {
	Food
}

func NewGoldenCarrot(identifier ItemIdentifier, name string) *GoldenCarrot {
	g := &GoldenCarrot{}
	g.Init(g, identifier, name)
	return g
}

func (g *GoldenCarrot) Clone() Item {
	cl := *g
	cl.rebind(&cl)
	return &cl
}

func (g *GoldenCarrot) GetFoodRestore() int { return 6 }

func (g *GoldenCarrot) GetSaturationRestore() float64 { return 14.4 }
