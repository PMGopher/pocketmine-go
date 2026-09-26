package io

import (
	"sync"

	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
)

// GlobalItemDataHandlers is a port of pocketmine\world\format\io\GlobalItemDataHandlers: the
// shared item serializer and deserializer. The item data upgrader isn't ported.
var (
	itemSerializer       *bedrockitem.ItemSerializer
	itemSerializerOnce   sync.Once
	itemDeserializer     *bedrockitem.ItemDeserializer
	itemDeserializerOnce sync.Once
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
