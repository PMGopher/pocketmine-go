package item

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// SplashPotion is a port of pocketmine\item\SplashPotion, a ProjectileItem (see item_use.go for
// throwing). Linger mirrors the
// constructor's private $linger flag, which PHP itself notes exists only for backward
// compatibility (LingeringPotion isn't a separate PHP class - it's just a SplashPotion
// constructed with linger=true via VanillaItems, which isn't ported either).
type SplashPotion struct {
	ItemBase

	PotionTypeValue PotionType
	Linger          bool
}

func NewSplashPotion(identifier ItemIdentifier, name string, linger bool, enchantmentTags ...string) *SplashPotion {
	s := &SplashPotion{PotionTypeValue: PotionTypeWater, Linger: linger}
	s.Init(s, identifier, name)
	s.enchantmentTags = enchantmentTags
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

// GetThrowForce is a port of SplashPotion::getThrowForce.
func (s *SplashPotion) GetThrowForce() float64 { return 0.5 }

// OnClickAir is ProjectileItem::onClickAir for splash potions.
func (s *SplashPotion) OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	return throwProjectile(s, player, directionVector)
}

// IsWaterPotion reports whether this is a splash water bottle (getType() === PotionType::WATER),
// see Potion.IsWaterPotion.
func (s *SplashPotion) IsWaterPotion() bool { return s.PotionTypeValue == PotionTypeWater }
