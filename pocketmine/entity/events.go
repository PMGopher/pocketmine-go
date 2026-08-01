package entity

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// DamagerBlock is the minimal surface EntityDamageByBlockEvent/EntityCombustByBlockEvent need
// from the block responsible for the damage/combustion — declared locally (rather than importing
// pocketmine-go/pocketmine/block) so this package doesn't depend on block at all. That keeps the
// door open for block to import entity later (to construct these events and call Entity.Attack
// for real, un-gapping the several OnEntityInside methods that currently stub that out) without
// an import cycle — block.Behavior already satisfies this trivially. Not wired into any block
// method yet; that's follow-up work, kept out of this initial slice to avoid touching the many
// existing block-package test doubles that would need a matching method added.
type DamagerBlock interface {
	GetName() string
}

// Call dispatches e to registered handlers on the global event Manager - a thin re-export of
// event.Call so entity.go doesn't need to spell out the generic type parameter at every call
// site. Equivalent to PHP's $event->call().
func Call[E any](e *E) { event.Call(e) }

// EntityLike is the interface an EntityEvent holds its subject as, instead of a concrete *Entity
// - needed so callers that only have an opaque interface value (like the block package's local
// Entity/Living interfaces) can still construct and fire these events; *Entity and *Living (and
// any future concrete mob type) satisfy it automatically.
type EntityLike interface {
	GetPosition() math.Vector3
	IsOnFire() bool
}

// DamageSource is the interface Entity.Attack accepts, rather than the concrete *EntityDamageEvent
// - satisfied by *EntityDamageEvent itself and by *EntityDamageByBlockEvent (which embeds it),
// letting Attack accept either polymorphically the way PHP's class hierarchy does natively. Call
// is part of this interface (rather than relying on struct embedding to promote one) because
// promoted methods keep the EMBEDDED type's receiver, not the outer type's - so
// EntityDamageByBlockEvent needs its own Call() to dispatch listeners registered for its own
// concrete type, not silently fire as a plain EntityDamageEvent (same self-dispatch pitfall this
// port has hit before in inventory.BaseInventory/CraftingGrid).
type DamageSource interface {
	GetCause() int
	GetFinalDamage() float64
	IsCancelled() bool
	Cancel()
	Call()
}

// EntityEvent is a port of pocketmine\event\entity\EntityEvent - embedded by every concrete event
// type below.
type EntityEvent struct {
	entity EntityLike
}

func (e *EntityEvent) GetEntity() EntityLike { return e.entity }

// EntityMotionEvent is a port of pocketmine\event\entity\EntityMotionEvent.
type EntityMotionEvent struct {
	EntityEvent
	event.CancellableTrait

	vector math.Vector3
}

func NewEntityMotionEvent(entity EntityLike, vector math.Vector3) *EntityMotionEvent {
	return &EntityMotionEvent{EntityEvent: EntityEvent{entity: entity}, vector: vector}
}

func (e *EntityMotionEvent) GetVector() math.Vector3 { return e.vector }

// EntityDamageEvent cause constants, a port of pocketmine\event\entity\EntityDamageEvent's
// CAUSE_* constants.
const (
	EntityDamageCauseContact = iota
	EntityDamageCauseEntityAttack
	EntityDamageCauseProjectile
	EntityDamageCauseSuffocation
	EntityDamageCauseFall
	EntityDamageCauseFire
	EntityDamageCauseFireTick
	EntityDamageCauseLava
	EntityDamageCauseDrowning
	EntityDamageCauseBlockExplosion
	EntityDamageCauseEntityExplosion
	EntityDamageCauseVoid
	EntityDamageCauseSuicide
	EntityDamageCauseMagic
	EntityDamageCauseCustom
	EntityDamageCauseStarvation
	EntityDamageCauseFallingBlock
)

// EntityDamageEvent damage modifier constants, a port of EntityDamageEvent's MODIFIER_*
// constants.
const (
	EntityDamageModifierArmor = iota + 1
	EntityDamageModifierStrength
	EntityDamageModifierWeakness
	EntityDamageModifierResistance
	EntityDamageModifierAbsorption
	EntityDamageModifierArmorEnchantments
	EntityDamageModifierCritical
	EntityDamageModifierTotem
	EntityDamageModifierWeaponEnchantments
	EntityDamageModifierPreviousDamageCooldown
	EntityDamageModifierArmorHelmet
)

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

// NewEntityDamageEvent is a port of EntityDamageEvent::__construct.
func NewEntityDamageEvent(entity EntityLike, cause int, damage float64, modifiers map[int]float64) *EntityDamageEvent {
	if modifiers == nil {
		modifiers = map[int]float64{}
	}
	originals := make(map[int]float64, len(modifiers))
	for k, v := range modifiers {
		originals[k] = v
	}
	return &EntityDamageEvent{
		EntityEvent:    EntityEvent{entity: entity},
		cause:          cause,
		baseDamage:     damage,
		originalBase:   damage,
		originals:      originals,
		modifiers:      modifiers,
		attackCooldown: 10,
	}
}

// Call dispatches this event to listeners registered for *EntityDamageEvent specifically - see
// DamageSource's doc comment for why this can't just be promoted.
func (e *EntityDamageEvent) Call() { event.Call(e) }

func (e *EntityDamageEvent) GetCause() int { return e.cause }

func (e *EntityDamageEvent) GetBaseDamage() float64 { return e.baseDamage }

func (e *EntityDamageEvent) SetBaseDamage(damage float64) { e.baseDamage = damage }

func (e *EntityDamageEvent) GetOriginalBaseDamage() float64 { return e.originalBase }

func (e *EntityDamageEvent) GetOriginalModifiers() map[int]float64 { return e.originals }

func (e *EntityDamageEvent) GetOriginalModifier(modifierType int) float64 {
	return e.originals[modifierType]
}

func (e *EntityDamageEvent) GetModifiers() map[int]float64 { return e.modifiers }

func (e *EntityDamageEvent) GetModifier(modifierType int) float64 { return e.modifiers[modifierType] }

func (e *EntityDamageEvent) SetModifier(damage float64, modifierType int) {
	if e.modifiers == nil {
		e.modifiers = map[int]float64{}
	}
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
	if sum < 0 {
		return 0
	}
	return sum
}

// CanBeReducedByArmor is a port of EntityDamageEvent::canBeReducedByArmor.
func (e *EntityDamageEvent) CanBeReducedByArmor() bool {
	switch e.cause {
	case EntityDamageCauseFireTick, EntityDamageCauseSuffocation, EntityDamageCauseDrowning,
		EntityDamageCauseStarvation, EntityDamageCauseFall, EntityDamageCauseVoid,
		EntityDamageCauseMagic, EntityDamageCauseSuicide:
		return false
	}
	return true
}

func (e *EntityDamageEvent) GetAttackCooldown() int { return e.attackCooldown }

func (e *EntityDamageEvent) SetAttackCooldown(attackCooldown int) { e.attackCooldown = attackCooldown }

// EntityDamageByBlockEvent is a port of pocketmine\event\entity\EntityDamageByBlockEvent.
type EntityDamageByBlockEvent struct {
	EntityDamageEvent

	damager DamagerBlock
}

func NewEntityDamageByBlockEvent(damager DamagerBlock, entity EntityLike, cause int, damage float64, modifiers map[int]float64) *EntityDamageByBlockEvent {
	return &EntityDamageByBlockEvent{
		EntityDamageEvent: *NewEntityDamageEvent(entity, cause, damage, modifiers),
		damager:           damager,
	}
}

// Call dispatches this event to listeners registered for *EntityDamageByBlockEvent specifically -
// see DamageSource's doc comment for why this can't just be promoted from EntityDamageEvent.
func (e *EntityDamageByBlockEvent) Call() { event.Call(e) }

func (e *EntityDamageByBlockEvent) GetDamager() DamagerBlock { return e.damager }

// EntityCombustEvent is a port of pocketmine\event\entity\EntityCombustEvent.
type EntityCombustEvent struct {
	EntityEvent
	event.CancellableTrait

	duration int
}

func NewEntityCombustEvent(combustee EntityLike, duration int) *EntityCombustEvent {
	return &EntityCombustEvent{EntityEvent: EntityEvent{entity: combustee}, duration: duration}
}

// Call dispatches this event to listeners registered for *EntityCombustEvent specifically - see
// DamageSource's doc comment for why EntityCombustByBlockEvent needs its own copy of this rather
// than inheriting it via embedding.
func (e *EntityCombustEvent) Call() { event.Call(e) }

func (e *EntityCombustEvent) GetDuration() int { return e.duration }

func (e *EntityCombustEvent) SetDuration(duration int) { e.duration = duration }

// EntityCombustByBlockEvent is a port of pocketmine\event\entity\EntityCombustByBlockEvent.
type EntityCombustByBlockEvent struct {
	EntityCombustEvent

	combuster DamagerBlock
}

func NewEntityCombustByBlockEvent(combuster DamagerBlock, combustee EntityLike, duration int) *EntityCombustByBlockEvent {
	return &EntityCombustByBlockEvent{
		EntityCombustEvent: *NewEntityCombustEvent(combustee, duration),
		combuster:          combuster,
	}
}

// Call dispatches this event to listeners registered for *EntityCombustByBlockEvent specifically.
func (e *EntityCombustByBlockEvent) Call() { event.Call(e) }

func (e *EntityCombustByBlockEvent) GetCombuster() DamagerBlock { return e.combuster }

// EntityExtinguishEvent cause constants, a port of EntityExtinguishEvent's CAUSE_* constants.
const (
	EntityExtinguishCauseCustom = iota
	EntityExtinguishCauseWater
	EntityExtinguishCauseWaterCauldron
	EntityExtinguishCauseRespawn
	EntityExtinguishCauseFireProof
	EntityExtinguishCauseTicking
	EntityExtinguishCauseRain
	EntityExtinguishCausePowderSnow
)

// EntityExtinguishEvent is a port of pocketmine\event\entity\EntityExtinguishEvent. Not
// cancellable, matching the PHP original (it doesn't use CancellableTrait).
type EntityExtinguishEvent struct {
	EntityEvent

	cause int
}

func NewEntityExtinguishEvent(entity EntityLike, cause int) *EntityExtinguishEvent {
	return &EntityExtinguishEvent{EntityEvent: EntityEvent{entity: entity}, cause: cause}
}

func (e *EntityExtinguishEvent) GetCause() int { return e.cause }
