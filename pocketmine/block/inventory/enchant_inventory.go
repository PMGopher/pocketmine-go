package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
)

const (
	EnchantSlotInput = 0
	EnchantSlotLapis = 1
)

// EnchantingViewer is the part of Player EnchantInventory needs: the enchantment seed and the
// network session's inventory manager.
type EnchantingViewer interface {
	playerevent.Player
	GetEnchantmentSeed() int
	GetInvManager() inventory.InventorySyncer
}

// EnchantingOptionsSyncer is InventoryManager::syncEnchantingTableOptions.
type EnchantingOptionsSyncer interface {
	SyncEnchantingTableOptions(options []*enchantment.EnchantingOption)
}

// EnchantInventory is a port of pocketmine\block\inventory\EnchantInventory.
type EnchantInventory struct {
	*inventory.SimpleInventory
	BlockInventoryTrait

	options []*enchantment.EnchantingOption
}

func NewEnchantInventory(holder block.Position) *EnchantInventory {
	e := &EnchantInventory{
		SimpleInventory:     inventory.NewSimpleInventory(2),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
	}
	// Dispatch onSlotChange (and listeners/viewers' $this) to the EnchantInventory.
	e.SimpleInventory.Init(e)
	return e
}

// OverrideSlotChange is a port of EnchantInventory::onSlotChange: when the input changes, the
// options are regenerated for every viewer and sent to it.
func (e *EnchantInventory) OverrideSlotChange(index int, before item.Item) {
	if index != EnchantSlotInput {
		return
	}
	for _, v := range e.GetViewers() {
		viewer, ok := v.(EnchantingViewer)
		if !ok {
			continue
		}
		e.options = nil
		it := e.GetInput()
		options := enchantment.GenerateOptions(e.tableSurroundings(), it, viewer.GetEnchantmentSeed())

		ev := playerevent.NewPlayerEnchantingOptionsRequestEvent(viewer, e, options)
		event.Call(ev)
		if !ev.IsCancelled() && len(ev.GetOptions()) > 0 {
			e.options = ev.GetOptions()
			if syncer, ok := viewer.GetInvManager().(EnchantingOptionsSyncer); ok {
				syncer.SyncEnchantingTableOptions(e.options)
			}
		}
	}
}

// tableSurroundings looks at the blocks around the holder for EnchantingHelper::countBookshelves.
func (e *EnchantInventory) tableSurroundings() enchantment.TableSurroundings {
	world, err := e.Holder.GetWorld()
	if err != nil {
		return func(int, int, int) (bool, bool) { return false, false }
	}
	x, y, z := e.Holder.FloorX(), e.Holder.FloorY(), e.Holder.FloorZ()
	return func(dx, dy, dz int) (bool, bool) {
		blk := world.GetBlockAt(x+dx, y+dy, z+dz)
		if blk == nil {
			return false, false
		}
		typeID := blk.GetTypeId()
		return typeID == block.AIR, typeID == block.BOOKSHELF
	}
}

func (e *EnchantInventory) GetInput() item.Item { return e.GetItem(EnchantSlotInput) }

func (e *EnchantInventory) GetLapis() item.Item { return e.GetItem(EnchantSlotLapis) }

// GetOutput is a port of EnchantInventory::getOutput: the input enchanted with the option's
// enchantments, or nil if there is no such option.
func (e *EnchantInventory) GetOutput(optionID int) item.Item {
	option := e.GetOption(optionID)
	if option == nil {
		return nil
	}
	return item.EnchantItem(e.GetInput(), option.GetEnchantments())
}

// GetOption is a port of EnchantInventory::getOption.
func (e *EnchantInventory) GetOption(optionID int) *enchantment.EnchantingOption {
	if optionID < 0 || optionID >= len(e.options) {
		return nil
	}
	return e.options[optionID]
}
