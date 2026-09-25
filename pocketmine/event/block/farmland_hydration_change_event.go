package block

import "pocketmine-go/pocketmine/event"

// FarmlandMaxWetness is Farmland::MAX_WETNESS.
const FarmlandMaxWetness = 7

// FarmlandHydrationChangeEvent is a port of pocketmine\event\block\FarmlandHydrationChangeEvent:
// called when farmland hydration is updated.
type FarmlandHydrationChangeEvent struct {
	BlockEvent
	event.CancellableTrait

	oldHydration, newHydration int
}

func NewFarmlandHydrationChangeEvent(block Block, oldHydration, newHydration int) *FarmlandHydrationChangeEvent {
	return &FarmlandHydrationChangeEvent{BlockEvent: BlockEvent{block: block}, oldHydration: oldHydration, newHydration: newHydration}
}

func (e *FarmlandHydrationChangeEvent) GetOldHydration() int { return e.oldHydration }

func (e *FarmlandHydrationChangeEvent) GetNewHydration() int { return e.newHydration }

func (e *FarmlandHydrationChangeEvent) SetNewHydration(hydration int) {
	if hydration < 0 || hydration > FarmlandMaxWetness {
		panic("Hydration must be in range 0 ... 7")
	}
	e.newHydration = hydration
}
