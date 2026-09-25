package player

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// PlayerEntityInteractEvent is a port of pocketmine\event\player\PlayerEntityInteractEvent:
// called when a player interacts with an entity (e.g. shearing a sheep, naming a mob, etc).
type PlayerEntityInteractEvent struct {
	PlayerEvent
	event.CancellableTrait

	entity   Entity
	clickPos math.Vector3
}

func NewPlayerEntityInteractEvent(player Player, entity Entity, clickPos math.Vector3) *PlayerEntityInteractEvent {
	return &PlayerEntityInteractEvent{PlayerEvent: PlayerEvent{player: player}, entity: entity, clickPos: clickPos}
}

func (e *PlayerEntityInteractEvent) GetEntity() Entity { return e.entity }

// GetClickPosition returns the absolute coordinates of the click. This is usually on the surface
// of the entity's hitbox.
func (e *PlayerEntityInteractEvent) GetClickPosition() math.Vector3 { return e.clickPos }
