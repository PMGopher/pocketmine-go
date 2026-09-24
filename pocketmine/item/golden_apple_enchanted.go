package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// GoldenAppleEnchanted is a port of pocketmine\item\GoldenAppleEnchanted.
type GoldenAppleEnchanted struct {
	GoldenApple
}

func NewGoldenAppleEnchanted(identifier ItemIdentifier, name string) *GoldenAppleEnchanted {
	g := &GoldenAppleEnchanted{}
	g.Init(g, identifier, name)
	return g
}

func (g *GoldenAppleEnchanted) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GoldenAppleEnchanted) GetAdditionalEffects() []*effect.EffectInstance {
	return []*effect.EffectInstance{
		effect.NewEffectInstanceWith(effect.VanillaRegeneration(), 600, 1),
		effect.NewEffectInstanceWith(effect.VanillaAbsorption(), 2400, 3),
		effect.NewEffectInstanceWith(effect.VanillaResistance(), 6000, 0),
		effect.NewEffectInstanceWith(effect.VanillaFireResistance(), 6000, 0),
	}
}
