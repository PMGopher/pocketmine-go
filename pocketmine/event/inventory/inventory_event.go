// Package inventory is a port of pocketmine\event\inventory: events about inventories,
// transactions, crafting and furnaces.
//
// Like the other event packages it sits below the packages that fire it (inventory, block/tile,
// player), so payloads are small local interfaces; listeners type-assert to the concrete type.
//
// Importers conventionally alias this package as inventoryevent.
package inventory

import (
	blockevent "pocketmine-go/pocketmine/event/block"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// Inventory is the surface these events need from pocketmine\inventory\Inventory.
type Inventory interface {
	GetSize() int
}

// Viewable is an Inventory that can list its viewers (Inventory::getViewers).
type Viewable interface {
	GetViewers() []any
}

// Player is the surface these events need from pocketmine\player\Player.
type Player = entityevent.Entity

// Item is the surface these events need from pocketmine\item\Item.
type Item = entityevent.Item

// InventoryEvent is a port of pocketmine\event\inventory\InventoryEvent.
type InventoryEvent struct {
	inventory Inventory
}

// NewInventoryEventBase builds the embedded InventoryEvent.
func NewInventoryEventBase(inventory Inventory) InventoryEvent {
	return InventoryEvent{inventory: inventory}
}

func (e *InventoryEvent) GetInventory() Inventory { return e.inventory }

// GetViewers is a port of InventoryEvent::getViewers.
func (e *InventoryEvent) GetViewers() []any {
	if v, ok := e.inventory.(Viewable); ok {
		return v.GetViewers()
	}
	return nil
}

// blockEventBase is blockevent.BlockEvent, for the furnace events that extend BlockEvent.
func blockEventBase(block blockevent.Block) blockevent.BlockEvent {
	return blockevent.NewBlockEventBase(block)
}
