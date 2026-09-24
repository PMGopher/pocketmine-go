package entity

import "pocketmine-go/pocketmine/event"

// ProjectileLaunchEvent is a port of pocketmine\event\entity\ProjectileLaunchEvent. The entity is a
// Projectile.
type ProjectileLaunchEvent struct {
	EntityEvent
	event.CancellableTrait
}

func NewProjectileLaunchEvent(entity Entity) *ProjectileLaunchEvent {
	return &ProjectileLaunchEvent{EntityEvent: EntityEvent{entity: entity}}
}

func (e *ProjectileLaunchEvent) Call() { event.Call(e) }
