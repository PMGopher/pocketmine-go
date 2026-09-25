package player

import "pocketmine-go/pocketmine/event"

// PlayerToggleFlightEvent is a port of pocketmine\event\player\PlayerToggleFlightEvent.
type PlayerToggleFlightEvent struct {
	PlayerEvent
	event.CancellableTrait

	isFlying bool
}

func NewPlayerToggleFlightEvent(player Player, isFlying bool) *PlayerToggleFlightEvent {
	return &PlayerToggleFlightEvent{PlayerEvent: PlayerEvent{player: player}, isFlying: isFlying}
}

func (e *PlayerToggleFlightEvent) IsFlying() bool { return e.isFlying }
