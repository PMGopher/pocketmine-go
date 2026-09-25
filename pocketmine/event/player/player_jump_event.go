package player

// PlayerJumpEvent is a port of pocketmine\event\player\PlayerJumpEvent: called when a player
// jumps.
type PlayerJumpEvent struct {
	PlayerEvent
}

func NewPlayerJumpEvent(player Player) *PlayerJumpEvent {
	return &PlayerJumpEvent{PlayerEvent: PlayerEvent{player: player}}
}
