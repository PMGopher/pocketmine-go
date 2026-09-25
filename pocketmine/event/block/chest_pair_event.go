package block

import "pocketmine-go/pocketmine/event"

// ChestPairEvent is a port of pocketmine\event\block\ChestPairEvent: called when two chests are
// about to be paired. left and right are pocketmine\block\Chest.
type ChestPairEvent struct {
	event.CancellableTrait

	left, right Block
}

func NewChestPairEvent(left, right Block) *ChestPairEvent {
	return &ChestPairEvent{left: left, right: right}
}

func (e *ChestPairEvent) GetLeft() Block  { return e.left }
func (e *ChestPairEvent) GetRight() Block { return e.right }
