package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

const (
	FurnaceSlotInput  = 0
	FurnaceSlotFuel   = 1
	FurnaceSlotResult = 2
)

// FurnaceInventory is a port of pocketmine\block\inventory\FurnaceInventory.
type FurnaceInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait

	furnaceType tile.FurnaceType
}

func NewFurnaceInventory(holder block.Position, furnaceType tile.FurnaceType) *FurnaceInventory {
	f := &FurnaceInventory{
		SimpleInventory:     inventory.NewSimpleInventory(3),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
		furnaceType:         furnaceType,
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	f.Init(f)
	return f
}

func (f *FurnaceInventory) GetFurnaceType() tile.FurnaceType { return f.furnaceType }

func (f *FurnaceInventory) GetResult() item.Item { return f.GetItem(FurnaceSlotResult) }

func (f *FurnaceInventory) GetFuel() item.Item { return f.GetItem(FurnaceSlotFuel) }

func (f *FurnaceInventory) GetSmelting() item.Item { return f.GetItem(FurnaceSlotInput) }

func (f *FurnaceInventory) SetResult(it item.Item) { f.SetItem(FurnaceSlotResult, it) }

func (f *FurnaceInventory) SetFuel(it item.Item) { f.SetItem(FurnaceSlotFuel, it) }

func (f *FurnaceInventory) SetSmelting(it item.Item) { f.SetItem(FurnaceSlotInput, it) }
