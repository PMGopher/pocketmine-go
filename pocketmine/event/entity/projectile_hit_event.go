package entity

import "pocketmine-go/pocketmine/math"

// ProjectileHitEvent is a port of the abstract pocketmine\event\entity\ProjectileHitEvent. The
// entity is a Projectile. Concrete hits are ProjectileHitBlockEvent/ProjectileHitEntityEvent;
// code that handles either uses the ProjectileHit interface.
type ProjectileHitEvent struct {
	EntityEvent

	rayTraceResult math.RayTraceResult
}

// ProjectileHit is the polymorphic view of a ProjectileHitEvent subclass (see DamageSource for why
// Go needs an interface here).
type ProjectileHit interface {
	GetEntity() Entity
	GetRayTraceResult() math.RayTraceResult
	Call()
}

// GetRayTraceResult returns a RayTraceResult object containing information such as the exact
// position struck, the AABB it hit, and the face of the AABB that it hit.
func (e *ProjectileHitEvent) GetRayTraceResult() math.RayTraceResult { return e.rayTraceResult }
