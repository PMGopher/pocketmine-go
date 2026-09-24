package item

import (
	"math/rand/v2"
	"pocketmine-go/pocketmine/entity/effect"
)

// RawChicken is a port of pocketmine\item\RawChicken.
type RawChicken struct {
	Food
}

func NewRawChicken(identifier ItemIdentifier, name string) *RawChicken {
	r := &RawChicken{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawChicken) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RawChicken) GetFoodRestore() int { return 2 }

func (r *RawChicken) GetSaturationRestore() float64 { return 1.2 }

// GetAdditionalEffects is a port of RawChicken::getAdditionalEffects (mt_rand(0, 9) < 3).
func (r *RawChicken) GetAdditionalEffects() []*effect.EffectInstance {
	if rand.IntN(10) < 3 {
		return []*effect.EffectInstance{effect.NewEffectInstanceWith(effect.VanillaHunger(), 600, 0)}
	}
	return nil
}
