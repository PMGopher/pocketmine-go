package player

// PlayerBedLeaveEvent is a port of pocketmine\event\player\PlayerBedLeaveEvent.
type PlayerBedLeaveEvent struct {
	PlayerEvent

	bed Block
}

func NewPlayerBedLeaveEvent(player Player, bed Block) *PlayerBedLeaveEvent {
	return &PlayerBedLeaveEvent{PlayerEvent: PlayerEvent{player: player}, bed: bed}
}

func (e *PlayerBedLeaveEvent) GetBed() Block { return e.bed }
