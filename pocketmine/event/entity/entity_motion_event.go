package entity

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// EntityMotionEvent is a port of pocketmine\event\entity\EntityMotionEvent.
type EntityMotionEvent struct {
	EntityEvent
	event.CancellableTrait

	mot math.Vector3
}

func NewEntityMotionEvent(entity Entity, mot math.Vector3) *EntityMotionEvent {
	return &EntityMotionEvent{EntityEvent: EntityEvent{entity: entity}, mot: mot}
}

func (e *EntityMotionEvent) Call() { event.Call(e) }

func (e *EntityMotionEvent) GetVector() math.Vector3 { return e.mot }
