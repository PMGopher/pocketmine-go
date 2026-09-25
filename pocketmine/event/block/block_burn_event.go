package block

import "pocketmine-go/pocketmine/event"

// BlockBurnEvent is a port of pocketmine\event\block\BlockBurnEvent: called when a block is
// burned away by fire.
type BlockBurnEvent struct {
	BlockEvent
	event.CancellableTrait

	causingBlock Block
}

func NewBlockBurnEvent(block, causingBlock Block) *BlockBurnEvent {
	return &BlockBurnEvent{BlockEvent: BlockEvent{block: block}, causingBlock: causingBlock}
}

// GetCausingBlock returns the block (usually Fire) which caused the target block to be burned
// away.
func (e *BlockBurnEvent) GetCausingBlock() Block { return e.causingBlock }
