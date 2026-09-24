package entity

import "pocketmine-go/pocketmine/event"

// EntityTrampleFarmlandEvent is a port of pocketmine\event\entity\EntityTrampleFarmlandEvent. PHP
// types the entity as Living; callers only construct it after an `instanceof Living` check.
type EntityTrampleFarmlandEvent struct {
	EntityEvent
	event.CancellableTrait

	block Block
}

func NewEntityTrampleFarmlandEvent(entity Entity, block Block) *EntityTrampleFarmlandEvent {
	return &EntityTrampleFarmlandEvent{EntityEvent: EntityEvent{entity: entity}, block: block}
}

func (e *EntityTrampleFarmlandEvent) Call() { event.Call(e) }

func (e *EntityTrampleFarmlandEvent) GetBlock() Block { return e.block }
