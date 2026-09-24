package entity

import "pocketmine-go/pocketmine/event"

// EntityDespawnEvent is a port of pocketmine\event\entity\EntityDespawnEvent - called when an
// entity is closed.
type EntityDespawnEvent struct {
	EntityEvent
}

func NewEntityDespawnEvent(entity Entity) *EntityDespawnEvent {
	return &EntityDespawnEvent{EntityEvent: EntityEvent{entity: entity}}
}

func (e *EntityDespawnEvent) Call() { event.Call(e) }
