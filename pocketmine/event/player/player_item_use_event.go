package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

// PlayerItemUseEvent is a port of pocketmine\event\player\PlayerItemUseEvent: called when a
// player uses its held item, for example when throwing a projectile.
type PlayerItemUseEvent struct {
	PlayerEvent
	event.CancellableTrait

	item            Item
	directionVector math.Vector3
}

func NewPlayerItemUseEvent(player Player, item Item, directionVector math.Vector3) *PlayerItemUseEvent {
	return &PlayerItemUseEvent{PlayerEvent: PlayerEvent{player: player}, item: item, directionVector: directionVector}
}

func (e *PlayerItemUseEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

// GetDirectionVector returns the direction the player is aiming when activating this item. Used
// for projectile direction.
func (e *PlayerItemUseEvent) GetDirectionVector() math.Vector3 { return e.directionVector }
