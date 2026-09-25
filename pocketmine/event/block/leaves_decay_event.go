package block

import "pocketmine-go/pocketmine/event"

// LeavesDecayEvent is a port of pocketmine\event\block\LeavesDecayEvent: called when leaves
// decay due to not being attached to wood.
type LeavesDecayEvent struct {
	BlockEvent
	event.CancellableTrait
}

func NewLeavesDecayEvent(block Block) *LeavesDecayEvent {
	return &LeavesDecayEvent{BlockEvent: BlockEvent{block: block}}
}
