package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// InstantHealthEffect is a port of pocketmine\entity\effect\InstantHealthEffect.
type InstantHealthEffect struct{ InstantEffect }

func NewInstantHealthEffect(name any, c color.Color, bad, hasBubbles bool) *InstantHealthEffect {
	e := &InstantHealthEffect{}
	e.initInstant(name, c, bad, hasBubbles)
	e.Init(e)
	return e
}

func (e *InstantHealthEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if entity.GetHealth() < float64(entity.GetMaxHealth()) {
		entity.Heal(entityevent.NewEntityRegainHealthEvent(entity, float64(int(4)<<instance.GetAmplifier())*potency, entityevent.RegainCauseMagic))
	}
}
