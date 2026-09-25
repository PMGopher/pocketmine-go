package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// PlayerExhaustEvent cause constants, a port of PlayerExhaustEvent's CAUSE_* constants.
const (
	ExhaustCauseAttack        = 1
	ExhaustCauseDamage        = 2
	ExhaustCauseMining        = 3
	ExhaustCauseHealthRegen   = 4
	ExhaustCausePotion        = 5
	ExhaustCauseWalking       = 6
	ExhaustCauseSprinting     = 7
	ExhaustCauseSwimming      = 8
	ExhaustCauseJumping       = 9
	ExhaustCauseSprintJumping = 10
	ExhaustCauseCustom        = 11
)

// PlayerExhaustEvent is a port of pocketmine\event\player\PlayerExhaustEvent. The entity is a
// Human.
type PlayerExhaustEvent struct {
	entityevent.EntityEvent
	event.CancellableTrait

	amount float64
	cause  int
}

func NewPlayerExhaustEvent(human entityevent.Entity, amount float64, cause int) *PlayerExhaustEvent {
	e := &PlayerExhaustEvent{amount: amount, cause: cause}
	e.EntityEvent = entityevent.NewEntityEventBase(human)
	return e
}

func (e *PlayerExhaustEvent) Call() { event.Call(e) }

// GetPlayer is a port of PlayerExhaustEvent::getPlayer (returning the Human).
func (e *PlayerExhaustEvent) GetPlayer() entityevent.Entity { return e.GetEntity() }

func (e *PlayerExhaustEvent) GetAmount() float64 { return e.amount }

func (e *PlayerExhaustEvent) SetAmount(amount float64) { e.amount = amount }

// GetCause returns an int cause of the exhaustion - one of the ExhaustCause* constants.
func (e *PlayerExhaustEvent) GetCause() int { return e.cause }
