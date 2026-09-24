package entity

import "pocketmine-go/pocketmine/event"

// EntityItemPickupEvent is a port of pocketmine\event\entity\EntityItemPickupEvent - called when an
// entity picks up an item, arrow, trident, etc. The entity is the collector.
type EntityItemPickupEvent struct {
	EntityEvent
	event.CancellableTrait

	origin    Entity
	item      Item
	inventory Inventory
}

// NewEntityItemPickupEvent is a port of EntityItemPickupEvent::__construct. inventory may be nil
// (PHP's ?Inventory).
func NewEntityItemPickupEvent(collector, origin Entity, item Item, inventory Inventory) *EntityItemPickupEvent {
	return &EntityItemPickupEvent{EntityEvent: EntityEvent{entity: collector}, origin: origin, item: item, inventory: inventory}
}

func (e *EntityItemPickupEvent) Call() { event.Call(e) }

// GetOrigin returns the entity being picked up.
func (e *EntityItemPickupEvent) GetOrigin() Entity { return e.origin }

// GetItem returns a copy of the item to be picked up (PHP's `clone $this->item`).
func (e *EntityItemPickupEvent) GetItem() Item { return CloneItem(e.item) }

// SetItem changes the item to be picked up. Does not affect the origin entity.
func (e *EntityItemPickupEvent) SetItem(item Item) { e.item = CloneItem(item) }

// GetInventory returns the inventory the item will be added to, or nil if it will be discarded.
func (e *EntityItemPickupEvent) GetInventory() Inventory { return e.inventory }

// SetInventory changes the inventory the item will be added to. nil means it will be discarded.
func (e *EntityItemPickupEvent) SetInventory(inventory Inventory) { e.inventory = inventory }
