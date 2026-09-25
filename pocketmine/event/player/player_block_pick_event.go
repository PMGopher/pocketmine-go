package player

import "pocketmine-go/pocketmine/event"

// PlayerBlockPickEvent is a port of pocketmine\event\player\PlayerBlockPickEvent: called when a
// player middle-clicks on a block to get an item in creative mode.
type PlayerBlockPickEvent struct {
	PlayerEvent
	event.CancellableTrait

	blockClicked Block
	resultItem   Item
}

func NewPlayerBlockPickEvent(player Player, blockClicked Block, resultItem Item) *PlayerBlockPickEvent {
	return &PlayerBlockPickEvent{PlayerEvent: PlayerEvent{player: player}, blockClicked: blockClicked, resultItem: resultItem}
}

func (e *PlayerBlockPickEvent) GetBlock() Block { return e.blockClicked }

func (e *PlayerBlockPickEvent) GetResultItem() Item { return e.resultItem }
