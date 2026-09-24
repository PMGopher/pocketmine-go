package entity

import "pocketmine-go/pocketmine/event"

// ItemSpawnEvent is a port of pocketmine\event\entity\ItemSpawnEvent. The entity is an ItemEntity.
type ItemSpawnEvent struct {
	EntityEvent
}

func NewItemSpawnEvent(item Entity) *ItemSpawnEvent {
	return &ItemSpawnEvent{EntityEvent: EntityEvent{entity: item}}
}

func (e *ItemSpawnEvent) Call() { event.Call(e) }
