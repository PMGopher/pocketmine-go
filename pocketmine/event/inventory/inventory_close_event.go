package inventory

// InventoryCloseEvent is a port of pocketmine\event\inventory\InventoryCloseEvent.
type InventoryCloseEvent struct {
	InventoryEvent

	who Player
}

func NewInventoryCloseEvent(inventory Inventory, who Player) *InventoryCloseEvent {
	return &InventoryCloseEvent{InventoryEvent: InventoryEvent{inventory: inventory}, who: who}
}

func (e *InventoryCloseEvent) GetPlayer() Player { return e.who }
