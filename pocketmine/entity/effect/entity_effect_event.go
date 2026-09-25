package effect

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// EntityEffectEvent is a port of pocketmine\event\entity\EntityEffectEvent.
//
// The three effect events live in this package rather than pocketmine/event/entity because they
// carry an *EffectInstance, and this package already imports event/entity (effects construct
// damage/heal events) - the reverse import would be a cycle.
type EntityEffectEvent struct {
	entityevent.EntityEvent
	event.CancellableTrait

	effect *EffectInstance
}

func (e *EntityEffectEvent) GetEffect() *EffectInstance { return e.effect }

// EntityEffectAddEvent is a port of pocketmine\event\entity\EntityEffectAddEvent - called when an
// effect is added to an Entity.
type EntityEffectAddEvent struct {
	EntityEffectEvent

	oldEffect *EffectInstance
}

// NewEntityEffectAddEvent is a port of EntityEffectAddEvent::__construct. oldEffect may be nil.
func NewEntityEffectAddEvent(entity entityevent.Entity, effect, oldEffect *EffectInstance) *EntityEffectAddEvent {
	return &EntityEffectAddEvent{
		EntityEffectEvent: EntityEffectEvent{EntityEvent: entityevent.NewEntityEventBase(entity), effect: effect},
		oldEffect:         oldEffect,
	}
}

func (e *EntityEffectAddEvent) Call() { event.Call(e) }

// WillModify returns whether the effect addition will replace an existing effect already applied
// to the entity.
func (e *EntityEffectAddEvent) WillModify() bool { return e.HasOldEffect() }

func (e *EntityEffectAddEvent) HasOldEffect() bool { return e.oldEffect != nil }

func (e *EntityEffectAddEvent) GetOldEffect() *EffectInstance { return e.oldEffect }

// EntityEffectRemoveEvent is a port of pocketmine\event\entity\EntityEffectRemoveEvent.
type EntityEffectRemoveEvent struct {
	EntityEffectEvent
}

func NewEntityEffectRemoveEvent(entity entityevent.Entity, effect *EffectInstance) *EntityEffectRemoveEvent {
	return &EntityEffectRemoveEvent{EntityEffectEvent: EntityEffectEvent{EntityEvent: entityevent.NewEntityEventBase(entity), effect: effect}}
}

func (e *EntityEffectRemoveEvent) Call() { event.Call(e) }

// Cancel is a port of EntityEffectRemoveEvent::cancel: removal of expired effects cannot be
// cancelled (panics like PHP's LogicException).
func (e *EntityEffectRemoveEvent) Cancel() {
	if e.GetEffect().GetDuration() <= 0 {
		panic("Removal of expired effects cannot be cancelled")
	}
	e.EntityEffectEvent.Cancel()
}

// EntityEffectEvent isn't abstract in PHP, so its handlers receive both subclasses.
func init() {
	event.DeclareParent[EntityEffectAddEvent, EntityEffectEvent]()
	event.DeclareParent[EntityEffectRemoveEvent, EntityEffectEvent]()
}
