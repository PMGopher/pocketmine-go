package io

import (
	"sync"

	"pocketmine-go/pocketmine/data"
	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
	itemupgrade "pocketmine-go/pocketmine/data/bedrock/item/upgrade"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// GlobalItemDataHandlers is a port of pocketmine\world\format\io\GlobalItemDataHandlers: the
// shared item serializer, deserializer and data upgrader.
var (
	itemSerializer       *bedrockitem.ItemSerializer
	itemSerializerOnce   sync.Once
	itemDeserializer     *bedrockitem.ItemDeserializer
	itemDeserializerOnce sync.Once
	itemDataUpgrader     *itemupgrade.ItemDataUpgrader
	itemDataUpgraderOnce sync.Once
)

// GetItemSerializer is GlobalItemDataHandlers::getSerializer.
func GetItemSerializer() *bedrockitem.ItemSerializer {
	itemSerializerOnce.Do(func() { itemSerializer = bedrockitem.NewItemSerializer(GetBlockStateSerializer()) })
	return itemSerializer
}

// GetItemDeserializer is GlobalItemDataHandlers::getDeserializer.
func GetItemDeserializer() *bedrockitem.ItemDeserializer {
	itemDeserializerOnce.Do(func() { itemDeserializer = bedrockitem.NewItemDeserializer(GetBlockStateDeserializer()) })
	return itemDeserializer
}

// GetItemDataUpgrader is GlobalItemDataHandlers::getUpgrader.
func GetItemDataUpgrader() *itemupgrade.ItemDataUpgrader {
	itemDataUpgraderOnce.Do(func() {
		u, err := itemupgrade.NewDefaultItemDataUpgrader(GetBlockDataUpgrader())
		if err != nil {
			// The data is embedded; this can only fail if it's corrupt.
			panic("loading item upgrade schemas: " + err.Error())
		}
		itemDataUpgrader = u
	})
	return itemDataUpgrader
}

// init gives the item package Item::nbtSerialize/nbtDeserialize, which go through these handlers
// (item can't import this package: it imports item).
func init() {
	item.NbtSerializeFunc = func(it item.Item, slot *int) (*nbt.CompoundTag, error) {
		stack, err := GetItemSerializer().SerializeStack(it, slot)
		if err != nil {
			return nil, err
		}
		return stack.ToNbt(), nil
	}
	item.NbtDeserializeFunc = func(tag *nbt.CompoundTag) (item.Item, error) {
		itemData, err := GetItemDataUpgrader().UpgradeItemStackNbt(tag)
		if err != nil {
			return nil, err
		}
		if itemData == nil {
			return item.VanillaAir(), nil
		}
		it, err := GetItemDeserializer().DeserializeStack(*itemData)
		if err != nil {
			return nil, &data.SavedDataLoadingError{Message: err.Error(), Cause: err}
		}
		return it, nil
	}
}
