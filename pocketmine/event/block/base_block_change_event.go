package block

import "pocketmine-go/pocketmine/event"

// BaseBlockChangeEvent is a port of pocketmine\event\block\BaseBlockChangeEvent: called when a
// block changes to another state.
type BaseBlockChangeEvent struct {
	BlockEvent
	event.CancellableTrait

	newState Block
}

// NewBaseBlockChangeEventBase builds the embedded BaseBlockChangeEvent.
func NewBaseBlockChangeEventBase(block, newState Block) BaseBlockChangeEvent {
	return BaseBlockChangeEvent{BlockEvent: BlockEvent{block: block}, newState: newState}
}

// GetNewState is a port of BaseBlockChangeEvent::getNewState.
func (e *BaseBlockChangeEvent) GetNewState() Block { return e.newState }
