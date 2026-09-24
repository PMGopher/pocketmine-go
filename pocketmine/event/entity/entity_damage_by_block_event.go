package entity

import "pocketmine-go/pocketmine/event"

// EntityDamageByBlockEvent is a port of pocketmine\event\entity\EntityDamageByBlockEvent - called
// when an entity takes damage from a block.
type EntityDamageByBlockEvent struct {
	EntityDamageEvent

	damager Block
}

// NewEntityDamageByBlockEvent is a port of EntityDamageByBlockEvent::__construct.
func NewEntityDamageByBlockEvent(damager Block, entity Entity, cause int, damage float64, modifiers map[int]float64) *EntityDamageByBlockEvent {
	e := &EntityDamageByBlockEvent{damager: damager}
	e.init(entity, cause, damage, modifiers)
	return e
}

// Call dispatches this event to listeners registered for *EntityDamageByBlockEvent - see
// DamageSource's doc comment for why this can't be promoted from EntityDamageEvent.
func (e *EntityDamageByBlockEvent) Call() { event.Call(e) }

func (e *EntityDamageByBlockEvent) GetDamager() Block { return e.damager }
