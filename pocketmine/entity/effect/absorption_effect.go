package effect

import "pocketmine-go/pocketmine/color"

// AbsorptionEffect is a port of pocketmine\entity\effect\AbsorptionEffect.
type AbsorptionEffect struct{ EffectBase }

func NewAbsorptionEffect(name any, c color.Color) *AbsorptionEffect {
	e := &AbsorptionEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *AbsorptionEffect) Add(entity Living, instance *EffectInstance) {
	newValue := float64(4 * instance.GetEffectLevel())
	if newValue > entity.GetAbsorption() {
		entity.SetAbsorption(newValue)
	}
}

func (e *AbsorptionEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetAbsorption(0)
}
