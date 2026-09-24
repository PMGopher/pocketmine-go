package entity

import "pocketmine-go/pocketmine/event"

// EntityShootBowEvent is a port of pocketmine\event\entity\EntityShootBowEvent. PHP types the
// shooter as Living and the projectile as Projectile.
type EntityShootBowEvent struct {
	EntityEvent
	event.CancellableTrait

	bow        Item
	projectile Entity
	force      float64
}

func NewEntityShootBowEvent(shooter Entity, bow Item, projectile Entity, force float64) *EntityShootBowEvent {
	return &EntityShootBowEvent{EntityEvent: EntityEvent{entity: shooter}, bow: bow, projectile: projectile, force: force}
}

func (e *EntityShootBowEvent) Call() { event.Call(e) }

func (e *EntityShootBowEvent) GetBow() Item { return e.bow }

// GetProjectile returns the entity considered as the projectile in this event.
func (e *EntityShootBowEvent) GetProjectile() Entity { return e.projectile }

// SetProjectile is a port of EntityShootBowEvent::setProjectile: a replaced projectile nobody has
// seen yet is closed, exactly like PHP.
func (e *EntityShootBowEvent) SetProjectile(projectile Entity) {
	if projectile != e.projectile {
		if old, ok := e.projectile.(interface {
			GetViewerCount() int
			Close()
		}); ok && old.GetViewerCount() == 0 {
			old.Close()
		}
		e.projectile = projectile
	}
}

func (e *EntityShootBowEvent) GetForce() float64 { return e.force }

func (e *EntityShootBowEvent) SetForce(force float64) { e.force = force }
