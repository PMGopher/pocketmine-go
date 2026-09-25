package inventory

import "pocketmine-go/pocketmine/event"

// InventoryOpenEvent is a port of pocketmine\event\inventory\InventoryOpenEvent.
type InventoryOpenEvent struct {
	InventoryEvent
	event.CancellableTrait

	who Player
}

func NewInventoryOpenEvent(inventory Inventory, who Player) *InventoryOpenEvent {
	return &InventoryOpenEvent{InventoryEvent: InventoryEvent{inventory: inventory}, who: who}
}

func (e *InventoryOpenEvent) GetPlayer() Player { return e.who }
