package block

// BlockSpreadEvent is a port of pocketmine\event\block\BlockSpreadEvent: called when a block
// spreads to another block, such as grass spreading to nearby dirt blocks.
type BlockSpreadEvent struct {
	BaseBlockChangeEvent

	source Block
}

func NewBlockSpreadEvent(block, source, newState Block) *BlockSpreadEvent {
	return &BlockSpreadEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState), source: source}
}

func (e *BlockSpreadEvent) GetSource() Block { return e.source }
