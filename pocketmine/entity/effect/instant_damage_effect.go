package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// InstantDamageEffect is a port of pocketmine\entity\effect\InstantDamageEffect.
type InstantDamageEffect struct{ InstantEffect }

func NewInstantDamageEffect(name any, c color.Color, bad, hasBubbles bool) *InstantDamageEffect {
	e := &InstantDamageEffect{}
	e.initInstant(name, c, bad, hasBubbles)
	e.Init(e)
	return e
}

func (e *InstantDamageEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	//TODO: add particles (witch spell)
	damage := float64(int(6)<<instance.GetAmplifier()) * potency
	var ev entityevent.DamageSource
	if source != nil {
		var sourceOwner entityevent.Entity
		if OwningEntityOf != nil {
			sourceOwner = OwningEntityOf(source)
		}
		if sourceOwner != nil {
			ev = entityevent.NewEntityDamageByChildEntityEvent(sourceOwner, source, entity, entityevent.CauseMagic, damage, nil)
		} else {
			ev = entityevent.NewEntityDamageByEntityEvent(source, entity, entityevent.CauseMagic, damage, nil)
		}
	} else {
		ev = entityevent.NewEntityDamageEvent(entity, entityevent.CauseMagic, damage, nil)
	}
	entity.Attack(ev)
}
