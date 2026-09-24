package effect

import "pocketmine-go/pocketmine/color"

// SlownessEffect is a port of pocketmine\entity\effect\SlownessEffect.
type SlownessEffect struct{ EffectBase }

func NewSlownessEffect(name any, c color.Color, bad bool) *SlownessEffect {
	e := &SlownessEffect{}
	e.initBase(name, c, bad, 600, true)
	e.Init(e)
	return e
}

func (e *SlownessEffect) Add(entity Living, instance *EffectInstance) {
	entity.SetMovementSpeed(entity.GetMovementSpeed()*(1-0.15*float64(instance.GetEffectLevel())), true)
}

func (e *SlownessEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetMovementSpeed(entity.GetMovementSpeed()/(1-0.15*float64(instance.GetEffectLevel())), false)
}
