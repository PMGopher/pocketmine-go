package entity

import "pocketmine-go/pocketmine/event"

// EntityCombustByBlockEvent is a port of pocketmine\event\entity\EntityCombustByBlockEvent.
type EntityCombustByBlockEvent struct {
	EntityCombustEvent

	combuster Block
}

func NewEntityCombustByBlockEvent(combuster Block, combustee Entity, duration int) *EntityCombustByBlockEvent {
	return &EntityCombustByBlockEvent{EntityCombustEvent: *NewEntityCombustEvent(combustee, duration), combuster: combuster}
}

func (e *EntityCombustByBlockEvent) Call() { event.Call(e) }

func (e *EntityCombustByBlockEvent) GetCombuster() Block { return e.combuster }
