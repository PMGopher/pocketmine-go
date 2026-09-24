package item

import (
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/utils"
)

// RottenFlesh is a port of pocketmine\item\RottenFlesh.
type RottenFlesh struct {
	Food
}

func NewRottenFlesh(identifier ItemIdentifier, name string) *RottenFlesh {
	r := &RottenFlesh{}
	r.Init(r, identifier, name)
	return r
}

func (r *RottenFlesh) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RottenFlesh) GetFoodRestore() int { return 4 }

func (r *RottenFlesh) GetSaturationRestore() float64 { return 0.8 }

func (r *RottenFlesh) GetAdditionalEffects() []*effect.EffectInstance {
	if utils.GetRandomFloat() <= 0.8 {
		return []*effect.EffectInstance{effect.NewEffectInstanceWith(effect.VanillaHunger(), 600, 0)}
	}
	return nil
}
