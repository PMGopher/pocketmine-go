package block

import "pocketmine-go/pocketmine/event"

// StructureGrowEvent is a port of pocketmine\event\block\StructureGrowEvent: called when
// structures such as Saplings or Bamboo grow. These types of plants tend to change multiple
// blocks at once upon growing.
type StructureGrowEvent struct {
	BlockEvent
	event.CancellableTrait

	transaction BlockTransaction
	player      Player
}

// NewStructureGrowEvent creates the event; player is nil when the structure grows by itself.
func NewStructureGrowEvent(block Block, transaction BlockTransaction, player Player) *StructureGrowEvent {
	return &StructureGrowEvent{BlockEvent: BlockEvent{block: block}, transaction: transaction, player: player}
}

func (e *StructureGrowEvent) GetTransaction() BlockTransaction { return e.transaction }

// GetPlayer returns the player which caused the structure to grow (for example by using bone
// meal), or nil.
func (e *StructureGrowEvent) GetPlayer() Player { return e.player }
