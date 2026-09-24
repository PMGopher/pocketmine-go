package effect

import "pocketmine-go/pocketmine/color"

// HealthBoostEffect is a port of pocketmine\entity\effect\HealthBoostEffect.
type HealthBoostEffect struct{ EffectBase }

func NewHealthBoostEffect(name any, c color.Color) *HealthBoostEffect {
	e := &HealthBoostEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *HealthBoostEffect) Add(entity Living, instance *EffectInstance) {
	entity.SetMaxHealth(entity.GetMaxHealth() + 4*instance.GetEffectLevel())
}

func (e *HealthBoostEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetMaxHealth(entity.GetMaxHealth() - 4*instance.GetEffectLevel())
	if entity.GetHealth() > float64(entity.GetMaxHealth()) {
		entity.SetHealth(float64(entity.GetMaxHealth()))
	}
}
