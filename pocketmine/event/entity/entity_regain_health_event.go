package entity

import "pocketmine-go/pocketmine/event"

// EntityRegainHealthEvent cause constants, a port of EntityRegainHealthEvent's CAUSE_* constants.
const (
	RegainCauseRegen      = 0
	RegainCauseEating     = 1
	RegainCauseMagic      = 2
	RegainCauseCustom     = 3
	RegainCauseSaturation = 4
)

// EntityRegainHealthEvent is a port of pocketmine\event\entity\EntityRegainHealthEvent.
type EntityRegainHealthEvent struct {
	EntityEvent
	event.CancellableTrait

	amount       float64
	regainReason int
}

func NewEntityRegainHealthEvent(entity Entity, amount float64, regainReason int) *EntityRegainHealthEvent {
	return &EntityRegainHealthEvent{EntityEvent: EntityEvent{entity: entity}, amount: amount, regainReason: regainReason}
}

func (e *EntityRegainHealthEvent) Call() { event.Call(e) }

func (e *EntityRegainHealthEvent) GetAmount() float64 { return e.amount }

func (e *EntityRegainHealthEvent) SetAmount(amount float64) { e.amount = amount }

// GetRegainReason returns one of the CAUSE_* constants to indicate why this regeneration occurred.
func (e *EntityRegainHealthEvent) GetRegainReason() int { return e.regainReason }
