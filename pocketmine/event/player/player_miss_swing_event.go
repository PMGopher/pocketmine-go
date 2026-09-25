package player

import "pocketmine-go/pocketmine/event"

// PlayerMissSwingEvent is a port of pocketmine\event\player\PlayerMissSwingEvent: called when a
// player attempts to perform the attack action (left-click) without a target entity.
type PlayerMissSwingEvent struct {
	PlayerEvent
	event.CancellableTrait
}

func NewPlayerMissSwingEvent(player Player) *PlayerMissSwingEvent {
	return &PlayerMissSwingEvent{PlayerEvent: PlayerEvent{player: player}}
}
