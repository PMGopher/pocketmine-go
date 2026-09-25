package player

import "pocketmine-go/pocketmine/event"

// PlayerToggleSwimEvent is a port of pocketmine\event\player\PlayerToggleSwimEvent.
type PlayerToggleSwimEvent struct {
	PlayerEvent
	event.CancellableTrait

	isSwimming bool
}

func NewPlayerToggleSwimEvent(player Player, isSwimming bool) *PlayerToggleSwimEvent {
	return &PlayerToggleSwimEvent{PlayerEvent: PlayerEvent{player: player}, isSwimming: isSwimming}
}

func (e *PlayerToggleSwimEvent) IsSwimming() bool { return e.isSwimming }
