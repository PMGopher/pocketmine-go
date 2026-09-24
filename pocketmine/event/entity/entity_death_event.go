package entity

import "pocketmine-go/pocketmine/event"

// EntityDeathEvent is a port of pocketmine\event\entity\EntityDeathEvent. PHP types the entity as
// Living.
type EntityDeathEvent struct {
	EntityEvent

	drops []Item
	xp    int
}

func NewEntityDeathEvent(entity Entity, drops []Item, xp int) *EntityDeathEvent {
	return &EntityDeathEvent{EntityEvent: EntityEvent{entity: entity}, drops: drops, xp: xp}
}

func (e *EntityDeathEvent) Call() { event.Call(e) }

func (e *EntityDeathEvent) GetDrops() []Item { return e.drops }

func (e *EntityDeathEvent) SetDrops(drops []Item) { e.drops = drops }

// GetXpDropAmount returns how much experience is dropped due to this entity's death.
func (e *EntityDeathEvent) GetXpDropAmount() int { return e.xp }

// SetXpDropAmount is a port of EntityDeathEvent::setXpDropAmount (panicking on a negative amount,
// like PHP's InvalidArgumentException).
func (e *EntityDeathEvent) SetXpDropAmount(xp int) {
	if xp < 0 {
		panic("XP drop amount must not be negative")
	}
	e.xp = xp
}
