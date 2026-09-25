package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// PlayerItemConsumeEvent is a port of pocketmine\event\player\PlayerItemConsumeEvent: called
// when a player eats something.
type PlayerItemConsumeEvent struct {
	PlayerEvent
	event.CancellableTrait

	item    Item
	residue []Item
}

func NewPlayerItemConsumeEvent(player Player, item Item, residue []Item) *PlayerItemConsumeEvent {
	return &PlayerItemConsumeEvent{PlayerEvent: PlayerEvent{player: player}, item: item, residue: residue}
}

func (e *PlayerItemConsumeEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

// GetResidue returns the items that will be added to the player's inventory after consuming the
// item.
func (e *PlayerItemConsumeEvent) GetResidue() []Item { return cloneItems(e.residue) }

// SetResidue sets the items that will be added to the player's inventory after consuming the
// item.
func (e *PlayerItemConsumeEvent) SetResidue(items []Item) { e.residue = cloneItems(items) }
