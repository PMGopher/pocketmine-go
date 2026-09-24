package entity

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// ProjectileHitBlockEvent is a port of pocketmine\event\entity\ProjectileHitBlockEvent.
type ProjectileHitBlockEvent struct {
	ProjectileHitEvent

	blockHit Block
}

func NewProjectileHitBlockEvent(entity Entity, rayTraceResult math.RayTraceResult, blockHit Block) *ProjectileHitBlockEvent {
	return &ProjectileHitBlockEvent{ProjectileHitEvent: ProjectileHitEvent{EntityEvent: EntityEvent{entity: entity}, rayTraceResult: rayTraceResult}, blockHit: blockHit}
}

func (e *ProjectileHitBlockEvent) Call() { event.Call(e) }

// GetBlockHit returns the Block struck by the projectile.
func (e *ProjectileHitBlockEvent) GetBlockHit() Block { return e.blockHit }
