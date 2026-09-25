package player

import "pocketmine-go/pocketmine/event"

// PlayerBedEnterEvent is a port of pocketmine\event\player\PlayerBedEnterEvent.
type PlayerBedEnterEvent struct {
	PlayerEvent
	event.CancellableTrait

	bed Block
}

func NewPlayerBedEnterEvent(player Player, bed Block) *PlayerBedEnterEvent {
	return &PlayerBedEnterEvent{PlayerEvent: PlayerEvent{player: player}, bed: bed}
}

func (e *PlayerBedEnterEvent) GetBed() Block { return e.bed }
