package entity

import "pocketmine-go/pocketmine/event"

// ItemDespawnEvent is a port of pocketmine\event\entity\ItemDespawnEvent. The entity is an
// ItemEntity.
type ItemDespawnEvent struct {
	EntityEvent
	event.CancellableTrait
}

func NewItemDespawnEvent(item Entity) *ItemDespawnEvent {
	return &ItemDespawnEvent{EntityEvent: EntityEvent{entity: item}}
}

func (e *ItemDespawnEvent) Call() { event.Call(e) }
