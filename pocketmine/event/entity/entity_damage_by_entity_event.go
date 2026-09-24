package entity

import "pocketmine-go/pocketmine/event"

// DefaultKnockbackForce/DefaultKnockbackVerticalLimit mirror Living::DEFAULT_KNOCKBACK_FORCE/
// DEFAULT_KNOCKBACK_VERTICAL_LIMIT - EntityDamageByEntityEvent's constructor defaults. Declared
// here (and re-exported by pocketmine/entity) since this package can't import entity.
const (
	DefaultKnockbackForce         = 0.4
	DefaultKnockbackVerticalLimit = 0.4
)

// AddAttackerModifiers is a port of EntityDamageByEntityEvent::addAttackerModifiers (the
// Strength/Weakness effect modifiers). The real body needs the damager's EffectManager and
// VanillaEffects, both of which live in pocketmine/entity/effect - a package that imports this one
// (effects construct damage events), so it can't be called from here directly. entity/effect's
// init() installs the real implementation; until then (e.g. in tests that never link the effect
// package) no attacker modifiers are applied.
var AddAttackerModifiers func(ev *EntityDamageByEntityEvent, damager Entity)

// EntityDamageByEntityEvent is a port of pocketmine\event\entity\EntityDamageByEntityEvent - called
// when an entity takes damage from another entity.
type EntityDamageByEntityEvent struct {
	EntityDamageEvent

	damager                Entity
	knockBack              float64
	verticalKnockBackLimit float64
}

// NewEntityDamageByEntityEvent is a port of EntityDamageByEntityEvent::__construct with the
// default knockBack/verticalKnockBackLimit (Living::DEFAULT_KNOCKBACK_FORCE/
// DEFAULT_KNOCKBACK_VERTICAL_LIMIT). Use NewEntityDamageByEntityEventWithKnockback for others.
func NewEntityDamageByEntityEvent(damager, entity Entity, cause int, damage float64, modifiers map[int]float64) *EntityDamageByEntityEvent {
	return NewEntityDamageByEntityEventWithKnockback(damager, entity, cause, damage, modifiers, DefaultKnockbackForce, DefaultKnockbackVerticalLimit)
}

// NewEntityDamageByEntityEventWithKnockback is the full port of EntityDamageByEntityEvent::__construct.
func NewEntityDamageByEntityEventWithKnockback(damager, entity Entity, cause int, damage float64, modifiers map[int]float64, knockBack, verticalKnockBackLimit float64) *EntityDamageByEntityEvent {
	e := &EntityDamageByEntityEvent{}
	e.initByEntity(damager, entity, cause, damage, modifiers, knockBack, verticalKnockBackLimit)
	return e
}

func (e *EntityDamageByEntityEvent) initByEntity(damager, entity Entity, cause int, damage float64, modifiers map[int]float64, knockBack, verticalKnockBackLimit float64) {
	e.damager = damager
	e.knockBack = knockBack
	e.verticalKnockBackLimit = verticalKnockBackLimit
	e.init(entity, cause, damage, modifiers)
	if AddAttackerModifiers != nil {
		AddAttackerModifiers(e, damager)
	}
}

// Call dispatches this event to listeners registered for *EntityDamageByEntityEvent.
func (e *EntityDamageByEntityEvent) Call() { event.Call(e) }

// GetDamager is a port of EntityDamageByEntityEvent::getDamager. PHP stores the damager's runtime
// ID and resolves it through WorldManager::findEntity, returning null once the damager no longer
// exists; returning nil for a closed damager is the same contract.
func (e *EntityDamageByEntityEvent) GetDamager() Entity {
	if e.damager == nil || e.damager.IsClosed() {
		return nil
	}
	return e.damager
}

func (e *EntityDamageByEntityEvent) GetKnockBack() float64 { return e.knockBack }

func (e *EntityDamageByEntityEvent) SetKnockBack(knockBack float64) { e.knockBack = knockBack }

// GetVerticalKnockBackLimit returns the maximum upwards velocity the victim may have after being
// knocked back.
func (e *EntityDamageByEntityEvent) GetVerticalKnockBackLimit() float64 {
	return e.verticalKnockBackLimit
}

func (e *EntityDamageByEntityEvent) SetVerticalKnockBackLimit(verticalKnockBackLimit float64) {
	e.verticalKnockBackLimit = verticalKnockBackLimit
}
