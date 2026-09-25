package block

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// BrewingStand is the surface these events need from pocketmine\block\tile\BrewingStand.
type BrewingStand interface {
	GetBlock() Block
}

// BrewItemEvent is a port of pocketmine\event\block\BrewItemEvent: called when a brewing stand
// brews an item. recipe is a pocketmine\crafting\BrewingRecipe.
type BrewItemEvent struct {
	BlockEvent
	event.CancellableTrait

	brewingStand BrewingStand
	slot         int
	input        Item
	result       Item
	recipe       any
}

func NewBrewItemEvent(brewingStand BrewingStand, slot int, input, result Item, recipe any) *BrewItemEvent {
	return &BrewItemEvent{BlockEvent: BlockEvent{block: brewingStand.GetBlock()}, brewingStand: brewingStand, slot: slot, input: input, result: result, recipe: recipe}
}

func (e *BrewItemEvent) GetBrewingStand() BrewingStand { return e.brewingStand }

// GetSlot returns which slot of the brewing stand's inventory the potion is in.
func (e *BrewItemEvent) GetSlot() int { return e.slot }

func (e *BrewItemEvent) GetInput() Item { return entityevent.CloneItem(e.input) }

func (e *BrewItemEvent) GetResult() Item { return entityevent.CloneItem(e.result) }

func (e *BrewItemEvent) SetResult(result Item) { e.result = entityevent.CloneItem(result) }

func (e *BrewItemEvent) GetRecipe() any { return e.recipe }
