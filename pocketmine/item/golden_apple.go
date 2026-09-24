package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// GoldenApple is a port of pocketmine\item\GoldenApple.
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

func (g *GoldenApple) GetAdditionalEffects() []*effect.EffectInstance {
	return []*effect.EffectInstance{
		effect.NewEffectInstanceWith(effect.VanillaRegeneration(), 100, 1),
		effect.NewEffectInstanceWith(effect.VanillaAbsorption(), 2400, 0),
	}
}
