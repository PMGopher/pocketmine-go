package player

// PlayerViewDistanceChangeEvent is a port of pocketmine\event\player\PlayerViewDistanceChangeEvent:
// called when a player requests a different viewing distance than the current one.
type PlayerViewDistanceChangeEvent struct {
	PlayerEvent

	oldDistance, newDistance int
}

func NewPlayerViewDistanceChangeEvent(player Player, oldDistance, newDistance int) *PlayerViewDistanceChangeEvent {
	return &PlayerViewDistanceChangeEvent{PlayerEvent: PlayerEvent{player: player}, oldDistance: oldDistance, newDistance: newDistance}
}

// GetNewDistance returns the new view radius, measured in chunks.
func (e *PlayerViewDistanceChangeEvent) GetNewDistance() int { return e.newDistance }

// GetOldDistance returns the old view radius, measured in chunks. A value of -1 means that the
// player has just connected and did not have a view distance before this event.
func (e *PlayerViewDistanceChangeEvent) GetOldDistance() int { return e.oldDistance }
