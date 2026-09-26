package tile

import "pocketmine-go/pocketmine/nbt"

// Block is the minimal surface a tile needs from a block (FlowerPot's plant): block.Behavior
// satisfies it.
type Block interface {
	GetStateId() int
}

// Hooks for the parts of tiles that need packages which import this one (item, block,
// world/format/io, network/mcpe/convert). Each is set from those packages' init(), like
// block.OpenWindowFunc. While unset (tests that don't import them), the tile skips that data.
var (
	// ItemNbtSerializeFunc is Item::nbtSerialize($slot) (slot -1: none).
	ItemNbtSerializeFunc func(it Item, slot int) (*nbt.CompoundTag, error)
	// ItemSafeNbtDeserializeFunc is Item::safeNbtDeserialize($tag, $errorLogContext). Air is
	// returned as nil.
	ItemSafeNbtDeserializeFunc func(tag *nbt.CompoundTag, errorLogContext string) Item

	// PotionItemInfoFunc returns Cauldron's POTION_CONTAINER_TYPE_* and potion type save ID
	// (PotionTypeIdMap) for a potion item.
	PotionItemInfoFunc func(it Item) (containerType int, potionID int)
	// NewPotionItemFunc builds VanillaItems::POTION/SPLASH_POTION/LINGERING_POTION()->setType()
	// for a container type and potion type save ID.
	NewPotionItemFunc func(containerType, potionID int) (Item, error)

	// BlockFromLegacyIdMetaFunc is GlobalBlockStateHandlers' upgrader (upgradeIntIdMeta) +
	// deserializer + RuntimeBlockStateRegistry::fromStateId.
	BlockFromLegacyIdMetaFunc func(id, meta int) (Block, error)
	// BlockFromNbtFunc is the same for a blockstate compound (upgradeBlockStateNbt).
	BlockFromNbtFunc func(tag *nbt.CompoundTag) (Block, error)
	// BlockToNbtFunc is GlobalBlockStateHandlers::getSerializer()->serialize($stateId)->toNbt().
	BlockToNbtFunc func(b Block) (*nbt.CompoundTag, error)
	// NetworkBlockStateNbtFunc is TypeConverter's
	// getBlockTranslator()->internalIdToNetworkStateData($stateId)->toNbt().
	NetworkBlockStateNbtFunc func(b Block) *nbt.CompoundTag
	// ItemToNetworkNbtFunc is TypeConverter::getInstance()->getItemTranslator()->toNetworkNbt().
	ItemToNetworkNbtFunc func(it Item) *nbt.CompoundTag
	// IsAirFunc is `$block instanceof Air`.
	IsAirFunc func(b Block) bool
)

// saveItem is Item::nbtSerialize for tiles: nil if the hook is unset or serialization fails.
func saveItem(it Item, slot int) *nbt.CompoundTag {
	if ItemNbtSerializeFunc == nil || it == nil {
		return nil
	}
	tag, err := ItemNbtSerializeFunc(it, slot)
	if err != nil {
		return nil
	}
	return tag
}

// loadItem is Item::safeNbtDeserialize for tiles: nil for air or if the hook is unset.
func loadItem(tag *nbt.CompoundTag, errorLogContext string) Item {
	if ItemSafeNbtDeserializeFunc == nil {
		return nil
	}
	return ItemSafeNbtDeserializeFunc(tag, errorLogContext)
}

// networkItemNbt is TypeConverter's toNetworkNbt for tiles: nil if the hook is unset.
func networkItemNbt(it Item) *nbt.CompoundTag {
	if ItemToNetworkNbtFunc == nil || it == nil {
		return nil
	}
	return ItemToNetworkNbtFunc(it)
}

// isNullItem is $item->isNull() for this package's minimal Item interface.
func isNullItem(it Item) bool {
	if it == nil {
		return true
	}
	n, ok := it.(interface{ IsNull() bool })
	return ok && n.IsNull()
}

// bookPageCount is count($book->getPages()) for a WritableBookBase.
func bookPageCount(book Item) int {
	if counter, ok := book.(interface{ GetPageCount() int }); ok {
		return counter.GetPageCount()
	}
	return 0
}
