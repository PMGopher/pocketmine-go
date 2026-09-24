package effect

import "pocketmine-go/pocketmine/color"

// InstantEffect is a port of the abstract pocketmine\entity\effect\InstantEffect: effects with a
// default duration of one tick, which apply every tick if forced to last longer.
//
// Go has no abstract classes, so "is this an InstantEffect" (PHP's instanceof, used by splash
// potions and area effect clouds) is the IsInstantEffect interface below.
type InstantEffect struct{ EffectBase }

// IsInstantEffect is the marker every InstantEffect subtype carries.
type IsInstantEffect interface {
	Effect
	isInstant()
}

func (e *InstantEffect) isInstant() {}

func (e *InstantEffect) initInstant(name any, c color.Color, bad, hasBubbles bool) {
	e.initBase(name, c, bad, 1, hasBubbles)
}

// GetApplyInterval: if forced to last longer than 1 tick, these apply every tick.
func (e *InstantEffect) GetApplyInterval(instance *EffectInstance) int { return 1 }

// IsInstant reports whether effectType is an InstantEffect (PHP's `instanceof InstantEffect`).
func IsInstant(effectType Effect) bool {
	_, ok := effectType.(IsInstantEffect)
	return ok
}
