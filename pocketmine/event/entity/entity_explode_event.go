package entity

import "pocketmine-go/pocketmine/event"

// EntityExplodeEvent is a port of pocketmine\event\entity\EntityExplodeEvent - called when an
// entity explodes, after the explosion's impact has been calculated. Blocks and ignitions are
// block.Behavior values in practice (see Block's doc comment).
type EntityExplodeEvent struct {
	EntityEvent
	event.CancellableTrait

	position  Position
	blocks    []Block
	yield     float64
	ignitions []Block
}

// NewEntityExplodeEvent is a port of EntityExplodeEvent::__construct (panicking on an out-of-range
// yield, like the PHP InvalidArgumentException).
func NewEntityExplodeEvent(entity Entity, position Position, blocks []Block, yield float64, ignitions []Block) *EntityExplodeEvent {
	checkYield(yield)
	return &EntityExplodeEvent{EntityEvent: EntityEvent{entity: entity}, position: position, blocks: blocks, yield: yield, ignitions: ignitions}
}

func checkYield(yield float64) {
	if yield < 0.0 || yield > 100.0 {
		panic("Yield must be in range 0.0 - 100.0")
	}
}

func (e *EntityExplodeEvent) Call() { event.Call(e) }

func (e *EntityExplodeEvent) GetPosition() Position { return e.position }

// GetBlockList returns the blocks destroyed by the explosion.
func (e *EntityExplodeEvent) GetBlockList() []Block { return e.blocks }

func (e *EntityExplodeEvent) SetBlockList(blocks []Block) { e.blocks = blocks }

// GetYield returns the percentage chance of drops from each block destroyed by the explosion.
func (e *EntityExplodeEvent) GetYield() float64 { return e.yield }

func (e *EntityExplodeEvent) SetYield(yield float64) {
	checkYield(yield)
	e.yield = yield
}

// SetIgnitions sets the blocks that will be set on fire by the explosion.
func (e *EntityExplodeEvent) SetIgnitions(ignitions []Block) { e.ignitions = ignitions }

func (e *EntityExplodeEvent) GetIgnitions() []Block { return e.ignitions }
