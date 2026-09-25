package block

import "pocketmine-go/pocketmine/event"

// BlockUpdateEvent is a port of pocketmine\event\block\BlockUpdateEvent: called when a block is
// updated because of a nearby block change.
type BlockUpdateEvent struct {
	BlockEvent
	event.CancellableTrait
}

func NewBlockUpdateEvent(block Block) *BlockUpdateEvent {
	return &BlockUpdateEvent{BlockEvent: BlockEvent{block: block}}
}
