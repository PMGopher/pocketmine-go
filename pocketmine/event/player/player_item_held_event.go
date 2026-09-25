package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// PlayerItemHeldEvent is a port of pocketmine\event\player\PlayerItemHeldEvent.
type PlayerItemHeldEvent struct {
	PlayerEvent
	event.CancellableTrait

	item       Item
	hotbarSlot int
}

func NewPlayerItemHeldEvent(player Player, item Item, hotbarSlot int) *PlayerItemHeldEvent {
	return &PlayerItemHeldEvent{PlayerEvent: PlayerEvent{player: player}, item: item, hotbarSlot: hotbarSlot}
}

// GetSlot returns the hotbar slot the player is attempting to hold. NOTE: This event is called
// BEFORE the slot is equipped server-side.
func (e *PlayerItemHeldEvent) GetSlot() int { return e.hotbarSlot }

func (e *PlayerItemHeldEvent) GetItem() Item { return entityevent.CloneItem(e.item) }
