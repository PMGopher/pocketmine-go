package player

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/event"
)

// Location is pocketmine\entity\Location as the move event carries it: a position plus yaw and
// pitch.
type Location struct {
	Position
	Yaw, Pitch float64
}

// PlayerMoveEvent is a port of pocketmine\event\player\PlayerMoveEvent.
type PlayerMoveEvent struct {
	PlayerEvent
	event.CancellableTrait

	from, to Location
}

func NewPlayerMoveEvent(player Player, from, to Location) *PlayerMoveEvent {
	return &PlayerMoveEvent{PlayerEvent: PlayerEvent{player: player}, from: from, to: to}
}

func (e *PlayerMoveEvent) GetFrom() Location { return e.from }

func (e *PlayerMoveEvent) GetTo() Location { return e.to }

// SetTo is a port of PlayerMoveEvent::setTo, including Utils::checkLocationNotInfOrNaN.
func (e *PlayerMoveEvent) SetTo(to Location) {
	for _, c := range []float64{to.X, to.Y, to.Z, to.Yaw, to.Pitch} {
		if stdmath.IsNaN(c) || stdmath.IsInf(c, 0) {
			panic(fmt.Sprintf("location %v contains NaN or infinite components", to))
		}
	}
	e.to = to
}
