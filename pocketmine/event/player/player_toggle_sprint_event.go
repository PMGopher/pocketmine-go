package player

import "pocketmine-go/pocketmine/event"

// PlayerToggleSprintEvent is a port of pocketmine\event\player\PlayerToggleSprintEvent.
type PlayerToggleSprintEvent struct {
	PlayerEvent
	event.CancellableTrait

	isSprinting bool
}

func NewPlayerToggleSprintEvent(player Player, isSprinting bool) *PlayerToggleSprintEvent {
	return &PlayerToggleSprintEvent{PlayerEvent: PlayerEvent{player: player}, isSprinting: isSprinting}
}

func (e *PlayerToggleSprintEvent) IsSprinting() bool { return e.isSprinting }
