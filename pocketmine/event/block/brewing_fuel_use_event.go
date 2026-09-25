package block

import "pocketmine-go/pocketmine/event"

// BrewingFuelUseEvent is a port of pocketmine\event\block\BrewingFuelUseEvent: called when a
// brewing stand consumes a new fuel item.
type BrewingFuelUseEvent struct {
	BlockEvent
	event.CancellableTrait

	fuelTime     int
	brewingStand BrewingStand
}

func NewBrewingFuelUseEvent(brewingStand BrewingStand) *BrewingFuelUseEvent {
	return &BrewingFuelUseEvent{BlockEvent: BlockEvent{block: brewingStand.GetBlock()}, fuelTime: 20, brewingStand: brewingStand}
}

func (e *BrewingFuelUseEvent) GetBrewingStand() BrewingStand { return e.brewingStand }

// GetFuelTime returns how many times the fuel can be used for potion brewing before it runs out.
func (e *BrewingFuelUseEvent) GetFuelTime() int { return e.fuelTime }

// SetFuelTime sets how many times the fuel can be used for potion brewing before it runs out.
func (e *BrewingFuelUseEvent) SetFuelTime(fuelTime int) {
	if fuelTime <= 0 {
		panic("Fuel time must be positive")
	}
	e.fuelTime = fuelTime
}
