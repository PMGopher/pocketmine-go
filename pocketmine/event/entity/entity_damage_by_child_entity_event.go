package entity

import "pocketmine-go/pocketmine/event"

// EntityDamageByChildEntityEvent is a port of
// pocketmine\event\entity\EntityDamageByChildEntityEvent - called when an entity takes damage
// from an entity sourced from another entity, for example being hit by a snowball thrown by a
// Player.
type EntityDamageByChildEntityEvent struct {
	EntityDamageByEntityEvent

	child Entity
}

// NewEntityDamageByChildEntityEvent is a port of EntityDamageByChildEntityEvent::__construct.
func NewEntityDamageByChildEntityEvent(damager, childEntity, entity Entity, cause int, damage float64, modifiers map[int]float64) *EntityDamageByChildEntityEvent {
	e := &EntityDamageByChildEntityEvent{child: childEntity}
	e.initByEntity(damager, entity, cause, damage, modifiers, DefaultKnockbackForce, DefaultKnockbackVerticalLimit)
	return e
}

// Call dispatches this event to listeners registered for *EntityDamageByChildEntityEvent.
func (e *EntityDamageByChildEntityEvent) Call() { event.Call(e) }

// GetChild is a port of EntityDamageByChildEntityEvent::getChild: the entity which caused the
// damage, or nil if it has been closed (see GetDamager's doc comment).
func (e *EntityDamageByChildEntityEvent) GetChild() Entity {
	if e.child == nil || e.child.IsClosed() {
		return nil
	}
	return e.child
}
