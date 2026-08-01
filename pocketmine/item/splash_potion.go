package item

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// SplashPotion is a port of pocketmine\item\SplashPotion. In PHP this extends ProjectileItem
// (which handles creating and launching the thrown entity) - ProjectileItem isn't ported (needs
// a real Player/Location/Entity), so this embeds ItemBase directly instead. Linger mirrors the
// constructor's private $linger flag, which PHP itself notes exists only for backward
// compatibility (LingeringPotion isn't a separate PHP class - it's just a SplashPotion
// constructed with linger=true via VanillaItems, which isn't ported either).
type SplashPotion struct {
	ItemBase

	PotionTypeValue PotionType
	Linger          bool
}

func NewSplashPotion(identifier ItemIdentifier, name string, linger bool) *SplashPotion {
	s := &SplashPotion{PotionTypeValue: PotionTypeWater, Linger: linger}
	s.Init(s, identifier, name)
	return s
}

func (s *SplashPotion) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SplashPotion) GetType() PotionType { return s.PotionTypeValue }

func (s *SplashPotion) SetType(t PotionType) { s.PotionTypeValue = t }

func (s *SplashPotion) GetMaxStackSize() int { return 1 }

func (s *SplashPotion) describeState(w runtime.DataDescriber) {
	t := int(s.PotionTypeValue)
	w.BoundedIntAuto(int(PotionTypeWater), int(PotionTypeStrongSlowness), &t)
	s.PotionTypeValue = PotionType(t)
}
