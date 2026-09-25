package block

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// CampfireCookEvent is a port of pocketmine\event\block\CampfireCookEvent. campfire is the
// pocketmine\block\Campfire.
type CampfireCookEvent struct {
	BlockEvent
	event.CancellableTrait

	campfire Block
	slot     int
	input    Item
	result   Item
}

func NewCampfireCookEvent(campfire Block, slot int, input, result Item) *CampfireCookEvent {
	return &CampfireCookEvent{BlockEvent: BlockEvent{block: campfire}, campfire: campfire, slot: slot, input: entityevent.CloneItem(input), result: result}
}

func (e *CampfireCookEvent) GetCampfire() Block { return e.campfire }

func (e *CampfireCookEvent) GetSlot() int { return e.slot }

func (e *CampfireCookEvent) GetInput() Item { return e.input }

func (e *CampfireCookEvent) GetResult() Item { return e.result }

func (e *CampfireCookEvent) SetResult(result Item) { e.result = result }
