package entity

import "pocketmine-go/pocketmine/event"

// EntityBlockChangeEvent is a port of pocketmine\event\entity\EntityBlockChangeEvent - called when
// an Entity, excluding players, changes a block directly (e.g. a FallingBlock landing).
type EntityBlockChangeEvent struct {
	EntityEvent
	event.CancellableTrait

	from Block
	to   Block
}

func NewEntityBlockChangeEvent(entity Entity, from, to Block) *EntityBlockChangeEvent {
	return &EntityBlockChangeEvent{EntityEvent: EntityEvent{entity: entity}, from: from, to: to}
}

func (e *EntityBlockChangeEvent) Call() { event.Call(e) }

func (e *EntityBlockChangeEvent) GetBlock() Block { return e.from }

func (e *EntityBlockChangeEvent) GetTo() Block { return e.to }
