package player

import "pocketmine-go/pocketmine/event"

// PlayerDropItemEvent is a port of pocketmine\event\player\PlayerDropItemEvent: called when a
// player tries to drop an item from its hotbar.
type PlayerDropItemEvent struct {
	PlayerEvent
	event.CancellableTrait

	drop Item
}

func NewPlayerDropItemEvent(player Player, drop Item) *PlayerDropItemEvent {
	return &PlayerDropItemEvent{PlayerEvent: PlayerEvent{player: player}, drop: drop}
}

func (e *PlayerDropItemEvent) GetItem() Item { return e.drop }
