package block

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// BlockItemPickupEvent is a port of pocketmine\event\block\BlockItemPickupEvent: called when a
// block picks up an item, arrow, etc (e.g. a hopper).
type BlockItemPickupEvent struct {
	BlockEvent
	event.CancellableTrait

	origin    Entity
	item      Item
	inventory Inventory
}

// NewBlockItemPickupEvent creates the event; inventory may be nil.
func NewBlockItemPickupEvent(collector Block, origin Entity, item Item, inventory Inventory) *BlockItemPickupEvent {
	return &BlockItemPickupEvent{BlockEvent: BlockEvent{block: collector}, origin: origin, item: item, inventory: inventory}
}

func (e *BlockItemPickupEvent) GetOrigin() Entity { return e.origin }

// GetItem returns (a clone of) the item to be collected.
func (e *BlockItemPickupEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

// SetItem changes the item to be collected.
func (e *BlockItemPickupEvent) SetItem(item Item) { e.item = entityevent.CloneItem(item) }

// GetInventory returns the inventory the item will be collected into, or nil if the item will be
// destroyed.
func (e *BlockItemPickupEvent) GetInventory() Inventory { return e.inventory }

// SetInventory changes the inventory the item will be collected into (nil destroys the item).
func (e *BlockItemPickupEvent) SetInventory(inventory Inventory) { e.inventory = inventory }
