package item

import (
	"math/rand/v2"
	"pocketmine-go/pocketmine/entity/effect"
)

// PoisonousPotato is a port of pocketmine\item\PoisonousPotato.
type PoisonousPotato struct {
	Food
}

func NewPoisonousPotato(identifier ItemIdentifier, name string) *PoisonousPotato {
	p := &PoisonousPotato{}
	p.Init(p, identifier, name)
	return p
}

func (p *PoisonousPotato) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PoisonousPotato) GetFoodRestore() int { return 2 }

func (p *PoisonousPotato) GetSaturationRestore() float64 { return 1.2 }

// GetAdditionalEffects is a port of PoisonousPotato::getAdditionalEffects (mt_rand(0, 100) > 40).
func (p *PoisonousPotato) GetAdditionalEffects() []*effect.EffectInstance {
	if rand.IntN(101) > 40 {
		return []*effect.EffectInstance{effect.NewEffectInstanceWith(effect.VanillaPoison(), 100, 0)}
	}
	return nil
}
