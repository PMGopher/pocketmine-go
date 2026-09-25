package player

import "pocketmine-go/pocketmine/event"

// PlayerDuplicateLoginEvent is a port of pocketmine\event\player\PlayerDuplicateLoginEvent:
// called when a player connects with a username or UUID that is already used by another player
// on the server. If cancelled, the newly connecting session will be disconnected; otherwise, the
// existing player will be disconnected.
type PlayerDuplicateLoginEvent struct {
	event.CancellableTrait
	PlayerDisconnectEventTrait

	connectingSession NetworkSession
	existingSession   NetworkSession
}

func NewPlayerDuplicateLoginEvent(connectingSession, existingSession NetworkSession, disconnectReason, disconnectScreenMessage any) *PlayerDuplicateLoginEvent {
	return &PlayerDuplicateLoginEvent{
		PlayerDisconnectEventTrait: PlayerDisconnectEventTrait{disconnectReason: disconnectReason, disconnectScreenMessage: disconnectScreenMessage},
		connectingSession:          connectingSession,
		existingSession:            existingSession,
	}
}

func (e *PlayerDuplicateLoginEvent) GetConnectingSession() NetworkSession { return e.connectingSession }

func (e *PlayerDuplicateLoginEvent) GetExistingSession() NetworkSession { return e.existingSession }
