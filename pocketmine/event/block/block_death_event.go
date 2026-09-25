package block

// BlockDeathEvent is a port of pocketmine\event\block\BlockDeathEvent: called when a block dies
// (e.g. a coral block drying out).
type BlockDeathEvent struct {
	BaseBlockChangeEvent
}

func NewBlockDeathEvent(block, newState Block) *BlockDeathEvent {
	return &BlockDeathEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState)}
}
