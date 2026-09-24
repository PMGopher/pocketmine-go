package entity

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// ProjectileHitEntityEvent is a port of pocketmine\event\entity\ProjectileHitEntityEvent.
type ProjectileHitEntityEvent struct {
	ProjectileHitEvent

	entityHit Entity
}

func NewProjectileHitEntityEvent(entity Entity, rayTraceResult math.RayTraceResult, entityHit Entity) *ProjectileHitEntityEvent {
	return &ProjectileHitEntityEvent{ProjectileHitEvent: ProjectileHitEvent{EntityEvent: EntityEvent{entity: entity}, rayTraceResult: rayTraceResult}, entityHit: entityHit}
}

func (e *ProjectileHitEntityEvent) Call() { event.Call(e) }

// GetEntityHit returns the Entity struck by the projectile.
func (e *ProjectileHitEntityEvent) GetEntityHit() Entity { return e.entityHit }
