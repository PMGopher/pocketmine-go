package player

import "pocketmine-go/pocketmine/event"

// PlayerToggleSneakEvent is a port of pocketmine\event\player\PlayerToggleSneakEvent.
type PlayerToggleSneakEvent struct {
	PlayerEvent
	event.CancellableTrait

	isSneaking     bool
	isSneakPressed bool
}

func NewPlayerToggleSneakEvent(player Player, isSneaking, isSneakPressed bool) *PlayerToggleSneakEvent {
	return &PlayerToggleSneakEvent{PlayerEvent: PlayerEvent{player: player}, isSneaking: isSneaking, isSneakPressed: isSneakPressed}
}

func (e *PlayerToggleSneakEvent) IsSneaking() bool { return e.isSneaking }

// IsSneakPressed returns whether the player is pressing the sneak key. The player may still be
// sneaking (e.g. in a 1.5 block high space) even if this is false.
func (e *PlayerToggleSneakEvent) IsSneakPressed() bool { return e.isSneakPressed }
