package player

import "pocketmine-go/pocketmine/event"

// PlayerKickEvent is a port of pocketmine\event\player\PlayerKickEvent: called when a player is
// kicked (forcibly disconnected) from the server, e.g. if an operator used /kick.
type PlayerKickEvent struct {
	PlayerEvent
	event.CancellableTrait
	PlayerDisconnectEventTrait

	quitMessage any
}

func NewPlayerKickEvent(player Player, disconnectReason, quitMessage, disconnectScreenMessage any) *PlayerKickEvent {
	return &PlayerKickEvent{
		PlayerEvent:                PlayerEvent{player: player},
		PlayerDisconnectEventTrait: PlayerDisconnectEventTrait{disconnectReason: disconnectReason, disconnectScreenMessage: disconnectScreenMessage},
		quitMessage:                quitMessage,
	}
}

// SetQuitMessage sets the quit message broadcasted to other players.
func (e *PlayerKickEvent) SetQuitMessage(quitMessage any) { e.quitMessage = quitMessage }

// GetQuitMessage returns the quit message broadcasted to other players.
func (e *PlayerKickEvent) GetQuitMessage() any { return e.quitMessage }
