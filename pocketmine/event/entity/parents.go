package entity

import "pocketmine-go/pocketmine/event"

// The handleable parent classes of this package's events (PHP `extends` of a non-abstract event
// or one tagged @allowHandle), so handlers of the parent also receive these (see
// event.DeclareParent).
func init() {
	event.DeclareParent[EntityDamageByEntityEvent, EntityDamageEvent]()
	event.DeclareParent[EntityDamageByBlockEvent, EntityDamageEvent]()
	event.DeclareParent[EntityDamageByChildEntityEvent, EntityDamageByEntityEvent]()
	event.DeclareParent[EntityCombustByBlockEvent, EntityCombustEvent]()
	event.DeclareParent[EntityCombustByEntityEvent, EntityCombustEvent]()
	event.DeclareParent[ProjectileHitBlockEvent, ProjectileHitEvent]()
	event.DeclareParent[ProjectileHitEntityEvent, ProjectileHitEvent]()
}
