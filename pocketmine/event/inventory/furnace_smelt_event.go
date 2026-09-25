package inventory

import (
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// FurnaceSmeltEvent is a port of pocketmine\event\inventory\FurnaceSmeltEvent.
type FurnaceSmeltEvent struct {
	blockevent.BlockEvent
	event.CancellableTrait

	furnace Furnace
	source  Item
	result  Item
}

// NewFurnaceSmeltEvent creates the event. source is copied with its count set to 1, like PHP.
func NewFurnaceSmeltEvent(furnace Furnace, source, result Item) *FurnaceSmeltEvent {
	src := entityevent.CloneItem(source)
	if c, ok := src.(interface{ SetCount(int) }); ok {
		c.SetCount(1)
	}
	return &FurnaceSmeltEvent{BlockEvent: blockEventBase(furnace.GetBlock()), furnace: furnace, source: src, result: result}
}

func (e *FurnaceSmeltEvent) GetFurnace() Furnace { return e.furnace }

func (e *FurnaceSmeltEvent) GetSource() Item { return e.source }

func (e *FurnaceSmeltEvent) GetResult() Item { return e.result }

func (e *FurnaceSmeltEvent) SetResult(result Item) { e.result = result }
