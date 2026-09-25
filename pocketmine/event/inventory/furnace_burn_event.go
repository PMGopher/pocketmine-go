package inventory

import (
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
)

// Furnace is the surface the furnace events need from pocketmine\block\tile\Furnace.
type Furnace interface {
	GetBlock() blockevent.Block
}

// FurnaceBurnEvent is a port of pocketmine\event\inventory\FurnaceBurnEvent: called when a
// furnace is about to consume a new fuel item.
type FurnaceBurnEvent struct {
	blockevent.BlockEvent
	event.CancellableTrait

	burning  bool
	furnace  Furnace
	fuel     Item
	burnTime int
}

func NewFurnaceBurnEvent(furnace Furnace, fuel Item, burnTime int) *FurnaceBurnEvent {
	return &FurnaceBurnEvent{BlockEvent: blockEventBase(furnace.GetBlock()), burning: true, furnace: furnace, fuel: fuel, burnTime: burnTime}
}

func (e *FurnaceBurnEvent) GetFurnace() Furnace { return e.furnace }

func (e *FurnaceBurnEvent) GetFuel() Item { return e.fuel }

// GetBurnTime returns the number of ticks that the furnace will be powered for.
func (e *FurnaceBurnEvent) GetBurnTime() int { return e.burnTime }

// SetBurnTime sets the number of ticks that the given fuel will power the furnace for.
func (e *FurnaceBurnEvent) SetBurnTime(burnTime int) { e.burnTime = burnTime }

// IsBurning returns whether the fuel item will be consumed.
func (e *FurnaceBurnEvent) IsBurning() bool { return e.burning }

// SetBurning sets whether the fuel will be consumed. If false, the furnace will smelt as if it
// consumed fuel, but the fuel item will not be deducted.
func (e *FurnaceBurnEvent) SetBurning(burning bool) { e.burning = burning }
