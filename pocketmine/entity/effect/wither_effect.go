package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// WitherEffect is a port of pocketmine\entity\effect\WitherEffect.
type WitherEffect struct{ EffectBase }

func NewWitherEffect(name any, c color.Color, bad bool) *WitherEffect {
	e := &WitherEffect{}
	e.initBase(name, c, bad, 600, true)
	e.Init(e)
	return e
}

func (e *WitherEffect) GetApplyInterval(instance *EffectInstance) int {
	return max(1, 40>>instance.GetAmplifier())
}

func (e *WitherEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	ev := entityevent.NewEntityDamageEvent(entity, entityevent.CauseMagic, 1, nil)
	entity.Attack(ev)
}
