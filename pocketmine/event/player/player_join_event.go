package player

// PlayerJoinEvent is a port of pocketmine\event\player\PlayerJoinEvent: called when the player
// spawns in the world after logging in, when they first see the terrain.
type PlayerJoinEvent struct {
	PlayerEvent

	joinMessage any
}

func NewPlayerJoinEvent(player Player, joinMessage any) *PlayerJoinEvent {
	return &PlayerJoinEvent{PlayerEvent: PlayerEvent{player: player}, joinMessage: joinMessage}
}

func (e *PlayerJoinEvent) SetJoinMessage(joinMessage any) { e.joinMessage = joinMessage }

func (e *PlayerJoinEvent) GetJoinMessage() any { return e.joinMessage }
