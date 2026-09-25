package block

import "pocketmine-go/pocketmine/event"

// BlockExplodeEvent is a port of pocketmine\event\block\BlockExplodeEvent: called when a block
// explodes, after the explosion's impact has been calculated.
type BlockExplodeEvent struct {
	BlockEvent
	event.CancellableTrait

	position  Position
	blocks    []Block
	yield     float64
	ignitions []Block
}

// NewBlockExplodeEvent creates the event. blocks are the blocks destroyed by the explosion,
// ignitions the positions that will be set on fire; yield is the percentage chance (0-100) of
// drops from each destroyed block.
func NewBlockExplodeEvent(block Block, position Position, blocks []Block, yield float64, ignitions []Block) *BlockExplodeEvent {
	checkYield(yield)
	return &BlockExplodeEvent{BlockEvent: BlockEvent{block: block}, position: position, blocks: blocks, yield: yield, ignitions: ignitions}
}

func checkYield(yield float64) {
	checkFinite("yield", yield)
	if yield < 0.0 || yield > 100.0 {
		panic("Yield must be in range 0.0 - 100.0")
	}
}

// GetPosition returns the center of the explosion.
func (e *BlockExplodeEvent) GetPosition() Position { return e.position }

// GetYield returns the percentage chance of drops from each block destroyed by the explosion.
func (e *BlockExplodeEvent) GetYield() float64 { return e.yield }

// SetYield sets the percentage chance of drops from each block destroyed by the explosion.
func (e *BlockExplodeEvent) SetYield(yield float64) {
	checkYield(yield)
	e.yield = yield
}

// GetAffectedBlocks returns a list of blocks destroyed by the explosion.
func (e *BlockExplodeEvent) GetAffectedBlocks() []Block { return e.blocks }

// SetAffectedBlocks sets the blocks destroyed by the explosion.
func (e *BlockExplodeEvent) SetAffectedBlocks(blocks []Block) { e.blocks = blocks }

// GetIgnitions returns a list of affected blocks that will be replaced by fire.
func (e *BlockExplodeEvent) GetIgnitions() []Block { return e.ignitions }

// SetIgnitions sets the blocks that will be replaced by fire.
func (e *BlockExplodeEvent) SetIgnitions(ignitions []Block) { e.ignitions = ignitions }
