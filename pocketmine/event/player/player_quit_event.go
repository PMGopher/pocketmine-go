package player

// PlayerQuitEvent is a port of pocketmine\event\player\PlayerQuitEvent: called when a player
// disconnects from the server for any reason. Some possible reasons include kicked, banned,
// connection lost or the client closing the game.
type PlayerQuitEvent struct {
	PlayerEvent

	quitMessage any
	quitReason  any
}

func NewPlayerQuitEvent(player Player, quitMessage, quitReason any) *PlayerQuitEvent {
	return &PlayerQuitEvent{PlayerEvent: PlayerEvent{player: player}, quitMessage: quitMessage, quitReason: quitReason}
}

func (e *PlayerQuitEvent) SetQuitMessage(quitMessage any) { e.quitMessage = quitMessage }

func (e *PlayerQuitEvent) GetQuitMessage() any { return e.quitMessage }

func (e *PlayerQuitEvent) GetQuitReason() any { return e.quitReason }
