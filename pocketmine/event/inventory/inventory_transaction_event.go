package inventory

import "pocketmine-go/pocketmine/event"

// InventoryTransactionEvent is a port of pocketmine\event\inventory\InventoryTransactionEvent:
// called when a player's inventory transaction is about to be executed. The transaction is an
// *transaction.InventoryTransaction.
type InventoryTransactionEvent struct {
	event.CancellableTrait

	transaction any
}

func NewInventoryTransactionEvent(transaction any) *InventoryTransactionEvent {
	return &InventoryTransactionEvent{transaction: transaction}
}

func (e *InventoryTransactionEvent) GetTransaction() any { return e.transaction }
