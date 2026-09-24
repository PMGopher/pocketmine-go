package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// PlayerExperienceChangeEvent is a port of pocketmine\event\player\PlayerExperienceChangeEvent -
// called when a player gains or loses XP levels and/or progress. The entity is a Human. A nil new
// level/progress means "unchanged" (PHP's ?int/?float).
type PlayerExperienceChangeEvent struct {
	entityevent.EntityEvent
	event.CancellableTrait

	oldLevel    int
	oldProgress float64
	newLevel    *int
	newProgress *float64
}

func NewPlayerExperienceChangeEvent(player entityevent.Entity, oldLevel int, oldProgress float64, newLevel *int, newProgress *float64) *PlayerExperienceChangeEvent {
	e := &PlayerExperienceChangeEvent{oldLevel: oldLevel, oldProgress: oldProgress, newLevel: newLevel, newProgress: newProgress}
	e.EntityEvent = entityevent.NewEntityEventBase(player)
	return e
}

func (e *PlayerExperienceChangeEvent) Call() { event.Call(e) }

func (e *PlayerExperienceChangeEvent) GetOldLevel() int { return e.oldLevel }

func (e *PlayerExperienceChangeEvent) GetOldProgress() float64 { return e.oldProgress }

func (e *PlayerExperienceChangeEvent) GetNewLevel() *int { return e.newLevel }

func (e *PlayerExperienceChangeEvent) GetNewProgress() *float64 { return e.newProgress }

func (e *PlayerExperienceChangeEvent) SetNewLevel(newLevel *int) { e.newLevel = newLevel }

// SetNewProgress is a port of PlayerExperienceChangeEvent::setNewProgress (panicking on an
// out-of-range value like the PHP InvalidArgumentException).
func (e *PlayerExperienceChangeEvent) SetNewProgress(newProgress *float64) {
	if newProgress != nil && (*newProgress < 0.0 || *newProgress > 1.0) {
		panic("XP progress must be in range 0-1")
	}
	e.newProgress = newProgress
}
