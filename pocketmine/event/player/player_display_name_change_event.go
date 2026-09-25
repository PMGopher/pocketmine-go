package player

// PlayerDisplayNameChangeEvent is a port of pocketmine\event\player\PlayerDisplayNameChangeEvent.
type PlayerDisplayNameChangeEvent struct {
	PlayerEvent

	oldName, newName string
}

func NewPlayerDisplayNameChangeEvent(player Player, oldName, newName string) *PlayerDisplayNameChangeEvent {
	return &PlayerDisplayNameChangeEvent{PlayerEvent: PlayerEvent{player: player}, oldName: oldName, newName: newName}
}

func (e *PlayerDisplayNameChangeEvent) GetOldName() string { return e.oldName }

func (e *PlayerDisplayNameChangeEvent) GetNewName() string { return e.newName }
