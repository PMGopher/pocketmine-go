package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// Pufferfish is a port of pocketmine\item\Pufferfish.
type Pufferfish struct {
	Food
}

func NewPufferfish(identifier ItemIdentifier, name string) *Pufferfish {
	p := &Pufferfish{}
	p.Init(p, identifier, name)
	return p
}

func (p *Pufferfish) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Pufferfish) GetFoodRestore() int { return 1 }

func (p *Pufferfish) GetSaturationRestore() float64 { return 0.2 }

func (p *Pufferfish) GetAdditionalEffects() []*effect.EffectInstance {
	return []*effect.EffectInstance{
		effect.NewEffectInstanceWith(effect.VanillaHunger(), 300, 2),
		effect.NewEffectInstanceWith(effect.VanillaPoison(), 1200, 3),
		effect.NewEffectInstanceWith(effect.VanillaNausea(), 300, 1),
	}
}
