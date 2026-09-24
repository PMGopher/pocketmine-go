package effect

import "pocketmine-go/pocketmine/color"

// InvisibilityEffect is a port of pocketmine\entity\effect\InvisibilityEffect.
type InvisibilityEffect struct{ EffectBase }

func NewInvisibilityEffect(name any, c color.Color) *InvisibilityEffect {
	e := &InvisibilityEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *InvisibilityEffect) Add(entity Living, instance *EffectInstance) {
	entity.SetInvisible(true)
	entity.SetNameTagVisible(false)
}

func (e *InvisibilityEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetInvisible(false)
	entity.SetNameTagVisible(true)
}
