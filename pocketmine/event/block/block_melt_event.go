package block

// BlockMeltEvent is a port of pocketmine\event\block\BlockMeltEvent: called when a block melts
// (ice, snow layers).
type BlockMeltEvent struct {
	BaseBlockChangeEvent
}

func NewBlockMeltEvent(block, newState Block) *BlockMeltEvent {
	return &BlockMeltEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState)}
}
