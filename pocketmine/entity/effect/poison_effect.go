package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// PoisonEffect is a port of pocketmine\entity\effect\PoisonEffect.
type PoisonEffect struct {
	EffectBase

	fatal bool
}

func NewPoisonEffect(name any, c color.Color, isBad bool, defaultDuration int, hasBubbles, fatal bool) *PoisonEffect {
	e := &PoisonEffect{fatal: fatal}
	e.initBase(name, c, isBad, defaultDuration, hasBubbles)
	e.Init(e)
	return e
}

func (e *PoisonEffect) GetApplyInterval(instance *EffectInstance) int {
	return max(1, 25>>instance.GetAmplifier())
}

func (e *PoisonEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if entity.GetHealth() > 1 || e.fatal {
		ev := entityevent.NewEntityDamageEvent(entity, entityevent.CauseMagic, 1, nil)
		entity.Attack(ev)
	}
}
