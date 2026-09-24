package entity

import "pocketmine-go/pocketmine/event"

// EntitySpawnEvent is a port of pocketmine\event\entity\EntitySpawnEvent - called when an entity
// is spawned (on its first tick).
type EntitySpawnEvent struct {
	EntityEvent
}

func NewEntitySpawnEvent(entity Entity) *EntitySpawnEvent {
	return &EntitySpawnEvent{EntityEvent: EntityEvent{entity: entity}}
}

func (e *EntitySpawnEvent) Call() { event.Call(e) }
