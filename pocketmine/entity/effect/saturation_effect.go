package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// SaturationEffect is a port of pocketmine\entity\effect\SaturationEffect.
type SaturationEffect struct{ InstantEffect }

func NewSaturationEffect(name any, c color.Color) *SaturationEffect {
	e := &SaturationEffect{}
	e.initInstant(name, c, false, true)
	e.Init(e)
	return e
}

func (e *SaturationEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if manager, ok := humanHungerManager(entity); ok {
		manager.AddFood(float64(instance.GetEffectLevel()))
		manager.AddSaturation(float64(instance.GetEffectLevel() * 2))
	}
}
