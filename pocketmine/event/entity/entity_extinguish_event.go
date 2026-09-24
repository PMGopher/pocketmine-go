package entity

import "pocketmine-go/pocketmine/event"

// EntityExtinguishEvent cause constants, a port of EntityExtinguishEvent's CAUSE_* constants.
const (
	ExtinguishCauseCustom        = 0
	ExtinguishCauseWater         = 1
	ExtinguishCauseWaterCauldron = 2
	ExtinguishCauseRespawn       = 3
	ExtinguishCauseFireProof     = 4
	ExtinguishCauseTicking       = 5
	ExtinguishCauseRain          = 6
	ExtinguishCausePowderSnow    = 7
)

// EntityExtinguishEvent is a port of pocketmine\event\entity\EntityExtinguishEvent. Not
// cancellable, matching the PHP original.
type EntityExtinguishEvent struct {
	EntityEvent

	cause int
}

func NewEntityExtinguishEvent(entity Entity, cause int) *EntityExtinguishEvent {
	return &EntityExtinguishEvent{EntityEvent: EntityEvent{entity: entity}, cause: cause}
}

func (e *EntityExtinguishEvent) Call() { event.Call(e) }

func (e *EntityExtinguishEvent) GetCause() int { return e.cause }
