package entity

import "pocketmine-go/pocketmine/event"

// EntityCombustEvent is a port of pocketmine\event\entity\EntityCombustEvent.
type EntityCombustEvent struct {
	EntityEvent
	event.CancellableTrait

	duration int
}

func NewEntityCombustEvent(combustee Entity, duration int) *EntityCombustEvent {
	return &EntityCombustEvent{EntityEvent: EntityEvent{entity: combustee}, duration: duration}
}

// Call dispatches this event to listeners registered for *EntityCombustEvent - the ByBlock/ByEntity
// subtypes each have their own Call for the same reason as DamageSource.Call.
func (e *EntityCombustEvent) Call() { event.Call(e) }

// GetDuration returns the duration in seconds the entity will burn for.
func (e *EntityCombustEvent) GetDuration() int { return e.duration }

func (e *EntityCombustEvent) SetDuration(duration int) { e.duration = duration }
