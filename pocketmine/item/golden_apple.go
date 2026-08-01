package item

// GoldenApple is a port of pocketmine\item\GoldenApple. GetAdditionalEffects (regeneration +
// absorption) isn't ported - EffectInstance (entity/effect package) isn't ported, same gap as
// Food.GetAdditionalEffects's doc comment.
type GoldenApple struct {
	Food
}

func NewGoldenApple(identifier ItemIdentifier, name string) *GoldenApple {
	g := &GoldenApple{}
	g.Init(g, identifier, name)
	return g
}

func (g *GoldenApple) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GoldenApple) RequiresHunger() bool { return false }

func (g *GoldenApple) GetFoodRestore() int { return 4 }

func (g *GoldenApple) GetSaturationRestore() float64 { return 9.6 }
