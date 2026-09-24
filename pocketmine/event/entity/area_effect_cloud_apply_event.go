package entity

import "pocketmine-go/pocketmine/event"

// AreaEffectCloudApplyEvent is a port of pocketmine\event\entity\AreaEffectCloudApplyEvent - called
// when an area effect cloud applies its effects. The entity is the AreaEffectCloud and every
// affected entity is a Living.
type AreaEffectCloudApplyEvent struct {
	EntityEvent
	event.CancellableTrait

	affectedEntities []Entity
}

func NewAreaEffectCloudApplyEvent(entity Entity, affectedEntities []Entity) *AreaEffectCloudApplyEvent {
	return &AreaEffectCloudApplyEvent{EntityEvent: EntityEvent{entity: entity}, affectedEntities: affectedEntities}
}

func (e *AreaEffectCloudApplyEvent) Call() { event.Call(e) }

// GetAffectedEntities returns the affected entities. PHP returns the array by value, so the result
// is a copy.
func (e *AreaEffectCloudApplyEvent) GetAffectedEntities() []Entity {
	return append([]Entity(nil), e.affectedEntities...)
}
