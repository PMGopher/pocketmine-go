package entity

import "pocketmine-go/pocketmine/event"

// ItemMergeEvent is a port of pocketmine\event\entity\ItemMergeEvent - called when a merge of two
// item entities is about to occur. Both entities are ItemEntity values.
type ItemMergeEvent struct {
	EntityEvent
	event.CancellableTrait

	target Entity
}

func NewItemMergeEvent(entity, target Entity) *ItemMergeEvent {
	return &ItemMergeEvent{EntityEvent: EntityEvent{entity: entity}, target: target}
}

func (e *ItemMergeEvent) Call() { event.Call(e) }

// GetTarget returns the merge destination.
func (e *ItemMergeEvent) GetTarget() Entity { return e.target }
