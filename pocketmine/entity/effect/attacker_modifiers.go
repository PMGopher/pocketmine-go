package effect

import entityevent "pocketmine-go/pocketmine/event/entity"

// effectHolder is the surface addAttackerModifiers needs from a damager to recognise it as a Living
// (PHP's `$damager instanceof Living` + getEffects()).
type effectHolder interface {
	GetEffects() *EffectManager
}

// addAttackerModifiers is a port of EntityDamageByEntityEvent::addAttackerModifiers - installed into
// entityevent.AddAttackerModifiers by init() below (see that variable's doc comment for why the body
// lives in this package).
func addAttackerModifiers(ev *entityevent.EntityDamageByEntityEvent, damager entityevent.Entity) {
	living, ok := damager.(effectHolder)
	if !ok { //TODO: move this to entity classes
		return
	}
	effects := living.GetEffects()
	if effects == nil {
		return
	}
	if strength := effects.Get(VanillaStrength()); strength != nil {
		ev.SetModifier(ev.GetBaseDamage()*0.3*float64(strength.GetEffectLevel()), entityevent.ModifierStrength)
	}
	if weakness := effects.Get(VanillaWeakness()); weakness != nil && ev.GetCause() == entityevent.CauseEntityAttack {
		ev.SetModifier(-(ev.GetBaseDamage() * 0.2 * float64(weakness.GetEffectLevel())), entityevent.ModifierWeakness)
	}
}

func init() {
	entityevent.AddAttackerModifiers = addAttackerModifiers
}
