package entity

import "pocketmine-go/pocketmine/event"

// EntityCombustByEntityEvent is a port of pocketmine\event\entity\EntityCombustByEntityEvent.
type EntityCombustByEntityEvent struct {
	EntityCombustEvent

	combuster Entity
}

func NewEntityCombustByEntityEvent(combuster, combustee Entity, duration int) *EntityCombustByEntityEvent {
	return &EntityCombustByEntityEvent{EntityCombustEvent: *NewEntityCombustEvent(combustee, duration), combuster: combuster}
}

func (e *EntityCombustByEntityEvent) Call() { event.Call(e) }

func (e *EntityCombustByEntityEvent) GetCombuster() Entity { return e.combuster }
