package block

// BlockFormEvent is a port of pocketmine\event\block\BlockFormEvent: called when a new block
// forms, usually as the result of some action (lava meeting water, snow falling).
type BlockFormEvent struct {
	BaseBlockChangeEvent

	causingBlock Block
}

func NewBlockFormEvent(block, newState, causingBlock Block) *BlockFormEvent {
	return &BlockFormEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState), causingBlock: causingBlock}
}

// GetCausingBlock returns the block which caused the target block to form into a new state.
func (e *BlockFormEvent) GetCausingBlock() Block { return e.causingBlock }
