package block

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// BlockTeleportEvent is a port of pocketmine\event\block\BlockTeleportEvent: called when a block
// (a dragon egg) teleports.
type BlockTeleportEvent struct {
	BlockEvent
	event.CancellableTrait

	to math.Vector3
}

func NewBlockTeleportEvent(block Block, to math.Vector3) *BlockTeleportEvent {
	return &BlockTeleportEvent{BlockEvent: BlockEvent{block: block}, to: to}
}

func (e *BlockTeleportEvent) GetTo() math.Vector3 { return e.to }

func (e *BlockTeleportEvent) SetTo(to math.Vector3) {
	checkVector3(to)
	e.to = to
}
