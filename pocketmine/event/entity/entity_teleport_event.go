package entity

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/event"
)

// EntityTeleportEvent is a port of pocketmine\event\entity\EntityTeleportEvent.
type EntityTeleportEvent struct {
	EntityEvent
	event.CancellableTrait

	from Position
	to   Position
}

func NewEntityTeleportEvent(entity Entity, from, to Position) *EntityTeleportEvent {
	return &EntityTeleportEvent{EntityEvent: EntityEvent{entity: entity}, from: from, to: to}
}

func (e *EntityTeleportEvent) Call() { event.Call(e) }

func (e *EntityTeleportEvent) GetFrom() Position { return e.from }

func (e *EntityTeleportEvent) GetTo() Position { return e.to }

// SetTo is a port of EntityTeleportEvent::setTo, including its Utils::checkVector3NotInfOrNaN
// validation (panicking, like the PHP InvalidArgumentException, on a programmer error).
func (e *EntityTeleportEvent) SetTo(to Position) {
	for _, c := range []float64{to.X, to.Y, to.Z} {
		if stdmath.IsNaN(c) || stdmath.IsInf(c, 0) {
			panic(fmt.Sprintf("entity: teleport target %v contains NaN or infinite components", to.Vector3))
		}
	}
	e.to = to
}
