package player

import "pocketmine-go/pocketmine/event"

// PlayerEntityPickEvent is a port of pocketmine\event\player\PlayerEntityPickEvent: called when
// a player middle-clicks on an entity to get an item in creative mode.
type PlayerEntityPickEvent struct {
	PlayerEvent
	event.CancellableTrait

	entityClicked Entity
	resultItem    Item
}

func NewPlayerEntityPickEvent(player Player, entityClicked Entity, resultItem Item) *PlayerEntityPickEvent {
	return &PlayerEntityPickEvent{PlayerEvent: PlayerEvent{player: player}, entityClicked: entityClicked, resultItem: resultItem}
}

func (e *PlayerEntityPickEvent) GetEntity() Entity { return e.entityClicked }

func (e *PlayerEntityPickEvent) GetResultItem() Item { return e.resultItem }
