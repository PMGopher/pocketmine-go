package player

import (
	"fmt"
	stdmath "math"
)

// PlayerRespawnEvent is a port of pocketmine\event\player\PlayerRespawnEvent: called when a
// player is respawned.
type PlayerRespawnEvent struct {
	PlayerEvent

	position Position
}

func NewPlayerRespawnEvent(player Player, position Position) *PlayerRespawnEvent {
	return &PlayerRespawnEvent{PlayerEvent: PlayerEvent{player: player}, position: position}
}

func (e *PlayerRespawnEvent) GetRespawnPosition() Position { return e.position }

// SetRespawnPosition is a port of PlayerRespawnEvent::setRespawnPosition: the position must
// reference a valid and loaded world.
func (e *PlayerRespawnEvent) SetRespawnPosition(position Position) {
	if position.World == nil {
		panic("Spawn position must reference a valid and loaded World")
	}
	for _, c := range []float64{position.X, position.Y, position.Z} {
		if stdmath.IsNaN(c) || stdmath.IsInf(c, 0) {
			panic(fmt.Sprintf("position %v contains NaN or infinite components", position.Vector3))
		}
	}
	e.position = position
}
