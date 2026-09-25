package player

import "pocketmine-go/pocketmine/event"

// PlayerEditBookEvent actions, PlayerEditBookEvent::ACTION_*.
const (
	EditBookActionReplacePage = 0
	EditBookActionAddPage     = 1
	EditBookActionDeletePage  = 2
	EditBookActionSwapPages   = 3
	EditBookActionSignBook    = 4
)

// PlayerEditBookEvent is a port of pocketmine\event\player\PlayerEditBookEvent. Books are
// item.WritableBookBase values.
type PlayerEditBookEvent struct {
	PlayerEvent
	event.CancellableTrait

	oldBook, newBook Item
	action           int
	modifiedPages    []int
}

func NewPlayerEditBookEvent(player Player, oldBook, newBook Item, action int, modifiedPages []int) *PlayerEditBookEvent {
	return &PlayerEditBookEvent{PlayerEvent: PlayerEvent{player: player}, oldBook: oldBook, newBook: newBook, action: action, modifiedPages: modifiedPages}
}

// GetAction returns the action of the event (one of the EditBookAction* constants).
func (e *PlayerEditBookEvent) GetAction() int { return e.action }

// GetOldBook returns the book before it was modified.
func (e *PlayerEditBookEvent) GetOldBook() Item { return e.oldBook }

// GetNewBook returns the book after it was modified. The new book may be a written book, if the
// book was signed.
func (e *PlayerEditBookEvent) GetNewBook() Item { return e.newBook }

// SetNewBook sets the new book as the result of this event.
func (e *PlayerEditBookEvent) SetNewBook(book Item) { e.newBook = book }

// GetModifiedPages returns an array containing the page IDs of modified pages.
func (e *PlayerEditBookEvent) GetModifiedPages() []int { return e.modifiedPages }
