package effect

import "pocketmine-go/pocketmine/color"

// SpeedEffect is a port of pocketmine\entity\effect\SpeedEffect.
type SpeedEffect struct{ EffectBase }

func NewSpeedEffect(name any, c color.Color) *SpeedEffect {
	e := &SpeedEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *SpeedEffect) Add(entity Living, instance *EffectInstance) {
	entity.SetMovementSpeed(entity.GetMovementSpeed()*(1+0.2*float64(instance.GetEffectLevel())), false)
}

func (e *SpeedEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetMovementSpeed(entity.GetMovementSpeed()/(1+0.2*float64(instance.GetEffectLevel())), false)
}
