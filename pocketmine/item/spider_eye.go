package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// SpiderEye is a port of pocketmine\item\SpiderEye.
type SpiderEye struct {
	Food
}

func NewSpiderEye(identifier ItemIdentifier, name string) *SpiderEye {
	s := &SpiderEye{}
	s.Init(s, identifier, name)
	return s
}

func (s *SpiderEye) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SpiderEye) GetFoodRestore() int { return 2 }

func (s *SpiderEye) GetSaturationRestore() float64 { return 3.2 }

func (s *SpiderEye) GetAdditionalEffects() []*effect.EffectInstance {
	return []*effect.EffectInstance{effect.NewEffectInstanceWith(effect.VanillaPoison(), 80, 0)}
}
