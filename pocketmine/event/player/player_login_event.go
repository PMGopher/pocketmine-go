package player

import "pocketmine-go/pocketmine/event"

// PlayerLoginEvent is a port of pocketmine\event\player\PlayerLoginEvent: called after the
// player has successfully authenticated, before it spawns. The player is on the loading screen
// when this is called. Cancelling this event will cause the player to be disconnected with the
// kick message set.
type PlayerLoginEvent struct {
	PlayerEvent
	event.CancellableTrait

	kickMessage any
}

func NewPlayerLoginEvent(player Player, kickMessage any) *PlayerLoginEvent {
	return &PlayerLoginEvent{PlayerEvent: PlayerEvent{player: player}, kickMessage: kickMessage}
}

func (e *PlayerLoginEvent) SetKickMessage(kickMessage any) { e.kickMessage = kickMessage }

func (e *PlayerLoginEvent) GetKickMessage() any { return e.kickMessage }
