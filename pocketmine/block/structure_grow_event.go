package block

import "pocketmine-go/pocketmine/event"

// StructureGrowEvent is a port of pocketmine\event\block\StructureGrowEvent. Called when
// structures such as Saplings or Bamboo grow - these tend to change multiple blocks at once,
// hence carrying a BlockTransactionImpl rather than a single before/after state like
// BlockChangeEvent's family.
type StructureGrowEvent struct {
	event.CancellableTrait

	Block       Behavior
	Transaction *BlockTransactionImpl
	Player      Player // nil when the structure grows by itself
}

func (e *StructureGrowEvent) GetBlock() Behavior { return e.Block }

func (e *StructureGrowEvent) GetTransaction() *BlockTransactionImpl { return e.Transaction }

func (e *StructureGrowEvent) GetPlayer() Player { return e.Player }
