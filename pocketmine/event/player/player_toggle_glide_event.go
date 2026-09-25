package player

import "pocketmine-go/pocketmine/event"

// PlayerToggleGlideEvent is a port of pocketmine\event\player\PlayerToggleGlideEvent.
type PlayerToggleGlideEvent struct {
	PlayerEvent
	event.CancellableTrait

	isGliding bool
}

func NewPlayerToggleGlideEvent(player Player, isGliding bool) *PlayerToggleGlideEvent {
	return &PlayerToggleGlideEvent{PlayerEvent: PlayerEvent{player: player}, isGliding: isGliding}
}

func (e *PlayerToggleGlideEvent) IsGliding() bool { return e.isGliding }
