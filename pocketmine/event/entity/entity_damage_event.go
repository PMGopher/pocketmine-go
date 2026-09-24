package entity

import (
	stdmath "math"

	"pocketmine-go/pocketmine/event"
)

// EntityDamageEvent modifier constants, a port of EntityDamageEvent's MODIFIER_* constants.
const (
	ModifierArmor                  = 1
	ModifierStrength               = 2
	ModifierWeakness               = 3
	ModifierResistance             = 4
	ModifierAbsorption             = 5
	ModifierArmorEnchantments      = 6
	ModifierCritical               = 7
	ModifierTotem                  = 8
	ModifierWeaponEnchantments     = 9
	ModifierPreviousDamageCooldown = 10
	ModifierArmorHelmet            = 11
)

// EntityDamageEvent cause constants, a port of EntityDamageEvent's CAUSE_* constants.
const (
	CauseContact         = 0
	CauseEntityAttack    = 1
	CauseProjectile      = 2
	CauseSuffocation     = 3
	CauseFall            = 4
	CauseFire            = 5
	CauseFireTick        = 6
	CauseLava            = 7
	CauseDrowning        = 8
	CauseBlockExplosion  = 9
	CauseEntityExplosion = 10
	CauseVoid            = 11
	CauseSuicide         = 12
	CauseMagic           = 13
	CauseCustom          = 14
	CauseStarvation      = 15
	CauseFallingBlock    = 16
)

// DamageSource is the polymorphic view of an EntityDamageEvent (or any subclass) - what
// Entity::attack() accepts. Go has no class hierarchy, so the embedding subclasses
// (EntityDamageByBlockEvent, EntityDamageByEntityEvent, EntityDamageByChildEntityEvent) are passed
// around through this interface and narrowed with a type switch wherever PHP uses instanceof.
//
// Call is part of the interface (rather than relying on embedding) because a promoted method keeps
// the embedded receiver: without its own Call, an EntityDamageByBlockEvent would dispatch to
// listeners registered for *EntityDamageEvent instead of its own type.
type DamageSource interface {
	GetEntity() Entity
	GetCause() int
	GetBaseDamage() float64
	SetBaseDamage(damage float64)
	GetOriginalBaseDamage() float64
	GetOriginalModifiers() map[int]float64
	GetOriginalModifier(modifierType int) float64
	GetModifiers() map[int]float64
	GetModifier(modifierType int) float64
	SetModifier(damage float64, modifierType int)
	IsApplicable(modifierType int) bool
	GetFinalDamage() float64
	CanBeReducedByArmor() bool
	GetAttackCooldown() int
	SetAttackCooldown(attackCooldown int)
	IsCancelled() bool
	Cancel()
	Uncancel()
	Call()
}

// EntityDamageEvent is a port of pocketmine\event\entity\EntityDamageEvent.
type EntityDamageEvent struct {
	EntityEvent
	event.CancellableTrait

	cause          int
	baseDamage     float64
	originalBase   float64
	originals      map[int]float64
	modifiers      map[int]float64
	attackCooldown int
}

var _ DamageSource = (*EntityDamageEvent)(nil)

// NewEntityDamageEvent is a port of EntityDamageEvent::__construct. modifiers may be nil (PHP's
// `array $modifiers = []` default).
func NewEntityDamageEvent(entity Entity, cause int, damage float64, modifiers map[int]float64) *EntityDamageEvent {
	e := &EntityDamageEvent{}
	e.init(entity, cause, damage, modifiers)
	return e
}

func (e *EntityDamageEvent) init(entity Entity, cause int, damage float64, modifiers map[int]float64) {
	current := make(map[int]float64, len(modifiers))
	originals := make(map[int]float64, len(modifiers))
	for k, v := range modifiers {
		current[k] = v
		originals[k] = v
	}
	e.entity = entity
	e.cause = cause
	e.baseDamage = damage
	e.originalBase = damage
	e.originals = originals
	e.modifiers = current
	e.attackCooldown = 10
}

// Call dispatches this event to listeners registered for *EntityDamageEvent.
func (e *EntityDamageEvent) Call() { event.Call(e) }

func (e *EntityDamageEvent) GetCause() int { return e.cause }

func (e *EntityDamageEvent) GetBaseDamage() float64 { return e.baseDamage }

// SetBaseDamage sets the base amount of damage applied, optionally recalculating modifiers.
func (e *EntityDamageEvent) SetBaseDamage(damage float64) { e.baseDamage = damage }

// GetOriginalBaseDamage returns the original base damage before any plugin modified it.
func (e *EntityDamageEvent) GetOriginalBaseDamage() float64 { return e.originalBase }

func (e *EntityDamageEvent) GetOriginalModifiers() map[int]float64 { return e.originals }

func (e *EntityDamageEvent) GetOriginalModifier(modifierType int) float64 {
	return e.originals[modifierType]
}

func (e *EntityDamageEvent) GetModifiers() map[int]float64 { return e.modifiers }

func (e *EntityDamageEvent) GetModifier(modifierType int) float64 { return e.modifiers[modifierType] }

func (e *EntityDamageEvent) SetModifier(damage float64, modifierType int) {
	e.modifiers[modifierType] = damage
}

func (e *EntityDamageEvent) IsApplicable(modifierType int) bool {
	_, ok := e.modifiers[modifierType]
	return ok
}

// GetFinalDamage is a port of EntityDamageEvent::getFinalDamage.
func (e *EntityDamageEvent) GetFinalDamage() float64 {
	sum := e.baseDamage
	for _, v := range e.modifiers {
		sum += v
	}
	return stdmath.Max(0, sum)
}

// CanBeReducedByArmor is a port of EntityDamageEvent::canBeReducedByArmor.
func (e *EntityDamageEvent) CanBeReducedByArmor() bool {
	switch e.cause {
	case CauseFireTick, CauseSuffocation, CauseDrowning, CauseStarvation, CauseFall, CauseVoid,
		CauseMagic, CauseSuicide:
		return false
	}
	return true
}

// GetAttackCooldown returns the cooldown in ticks before the target entity can be attacked again.
func (e *EntityDamageEvent) GetAttackCooldown() int { return e.attackCooldown }

// SetAttackCooldown sets the cooldown in ticks before the target entity can be attacked again.
// NOTE: This value is not used in non-Living entities.
func (e *EntityDamageEvent) SetAttackCooldown(attackCooldown int) { e.attackCooldown = attackCooldown }
