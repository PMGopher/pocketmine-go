package item

import (
	"errors"
	"fmt"

	"pocketmine-go/pocketmine/block/tile"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
)

// NbtSerializeFunc and NbtDeserializeFunc are GlobalItemDataHandlers' serializer and
// upgrader+deserializer, which Item::nbtSerialize/nbtDeserialize go through. This package can't
// import world/format/io (it imports item), so that package sets them in its init().
var (
	NbtSerializeFunc   func(it Item, slot *int) (*nbt.CompoundTag, error)
	NbtDeserializeFunc func(tag *nbt.CompoundTag) (Item, error)
)

var errNoItemDataHandlers = errors.New("item NBT handlers aren't loaded (import world/format/io)")

// NbtSerialize is a port of Item::nbtSerialize: the item as saved in inventories, chunks and
// player data. slot is -1 for a stack that isn't in a slot.
func NbtSerialize(it Item, slot int) (*nbt.CompoundTag, error) {
	if NbtSerializeFunc == nil {
		return nil, errNoItemDataHandlers
	}
	var slotPtr *int
	if slot != -1 {
		slotPtr = &slot
	}
	return NbtSerializeFunc(it, slotPtr)
}

// NbtDeserialize is a port of Item::nbtDeserialize: the item saved in tag, upgraded from older
// formats if needed (a *data.SavedDataLoadingError on failure).
func NbtDeserialize(tag *nbt.CompoundTag) (Item, error) {
	if NbtDeserializeFunc == nil {
		return nil, errNoItemDataHandlers
	}
	return NbtDeserializeFunc(tag)
}

// SafeNbtDeserialize is a port of Item::safeNbtDeserialize: like NbtDeserialize, but data errors
// are logged (to logger, or the global logger if nil) and the item is replaced by air.
// errorLogContext describes where the item was (inventory owner, slot, ...).
func SafeNbtDeserialize(tag *nbt.CompoundTag, errorLogContext string, logger log.Logger) Item {
	it, err := NbtDeserialize(tag)
	if err != nil {
		//TODO: what if the intention was to suppress logging?
		if logger == nil {
			logger = log.Global()
		}
		logger.Error(fmt.Sprintf("%s: Error deserializing item (item will be replaced by AIR): %v", errorLogContext, err))
		//no trace here, otherwise things could get very noisy
		return VanillaAir()
	}
	return it
}

// init gives the tile package the item parts of tiles (it can't import this package).
func init() {
	tile.ItemNbtSerializeFunc = func(it tile.Item, slot int) (*nbt.CompoundTag, error) {
		item, ok := it.(Item)
		if !ok {
			return nil, fmt.Errorf("%T is not an item", it)
		}
		return NbtSerialize(item, slot)
	}
	tile.ItemSafeNbtDeserializeFunc = func(tag *nbt.CompoundTag, errorLogContext string) tile.Item {
		it := SafeNbtDeserialize(tag, errorLogContext, nil)
		if it.IsNull() {
			return nil
		}
		return it
	}
	// Cauldron::addAdditionalSpawnData/writeSaveData's potion container type and PotionTypeIdMap ID.
	tile.PotionItemInfoFunc = func(it tile.Item) (int, int) {
		item, ok := it.(Item)
		if !ok {
			return tile.CauldronPotionContainerTypeNone, -1
		}
		containerType := tile.CauldronPotionContainerTypeNone
		switch item.GetTypeId() {
		case POTION:
			containerType = tile.CauldronPotionContainerTypeNormal
		case SPLASH_POTION:
			containerType = tile.CauldronPotionContainerTypeSplash
		case LINGERING_POTION:
			containerType = tile.CauldronPotionContainerTypeLingering
		default:
			panic("Unexpected potion item type")
		}
		potionID := -1
		if typed, ok := item.(interface{ GetType() PotionType }); ok {
			potionID = PotionTypeIdMapInstance.ToID(typed.GetType())
		}
		return containerType, potionID
	}
	// Cauldron::readSaveData's potion item.
	tile.NewPotionItemFunc = func(containerType, potionID int) (tile.Item, error) {
		potionType, ok := PotionTypeIdMapInstance.FromID(potionID)
		if !ok {
			return nil, fmt.Errorf("Unknown potion type ID %d", potionID)
		}
		var potion Item
		switch containerType {
		case tile.CauldronPotionContainerTypeNormal:
			potion = VanillaPotion()
		case tile.CauldronPotionContainerTypeSplash:
			potion = VanillaSplashPotion()
		case tile.CauldronPotionContainerTypeLingering:
			potion = VanillaLingeringPotion()
		default:
			return nil, tile.CauldronContainerTypeError(containerType)
		}
		potion.(interface{ SetType(PotionType) }).SetType(potionType)
		return potion, nil
	}
}
