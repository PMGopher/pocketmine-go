package convert

import (
	"fmt"
	"pocketmine-go/pocketmine/nbt"

	"pocketmine-go/pocketmine/data/bedrock"
	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
	"pocketmine-go/pocketmine/item"
	worldio "pocketmine-go/pocketmine/world/format/io"
)

// noBlockRuntimeID mirrors ItemTranslator::NO_BLOCK_RUNTIME_ID - a technically-valid block runtime
// ID (0) used to mean "this item isn't a blockitem".
const noBlockRuntimeID = 0

// ItemTranslator is a port of pocketmine\network\mcpe\convert\ItemTranslator: items to and from
// network item IDs, through the item serializer/deserializer (GlobalItemDataHandlers), the vendored
// item and block state dictionaries and BlockItemIdMap.
type ItemTranslator struct {
	serializer   *bedrockitem.ItemSerializer
	deserializer *bedrockitem.ItemDeserializer
	blockItemIDs *bedrockitem.BlockItemIdMap
}

func NewItemTranslator() *ItemTranslator {
	return &ItemTranslator{
		serializer:   worldio.GetItemSerializer(),
		deserializer: worldio.GetItemDeserializer(),
		blockItemIDs: bedrockitem.GetBlockItemIdMap(),
	}
}

// ToNetworkIDQuiet is a port of ItemTranslator::toNetworkIdQuiet: ok is false for an item type
// with no serializer.
func (t *ItemTranslator) ToNetworkIDQuiet(it item.Item) (networkID int32, meta int16, blockRuntimeID int32, ok bool) {
	networkID, meta, blockRuntimeID, err := t.ToNetworkID(it)
	return networkID, meta, blockRuntimeID, err == nil
}

// ToNetworkID is a port of ItemTranslator::toNetworkId: the network item ID, meta and block runtime
// ID (noBlockRuntimeID for items that aren't block items).
func (t *ItemTranslator) ToNetworkID(it item.Item) (networkID int32, meta int16, blockRuntimeID int32, err error) {
	//TODO: we should probably come up with a cache for this
	itemData, err := t.serializer.SerializeType(it)
	if err != nil {
		return 0, 0, 0, err
	}

	numericID, ok := bedrock.ItemRuntimeIDFor(itemData.Name)
	if !ok {
		return 0, 0, 0, &bedrockitem.ItemTypeSerializeError{Message: fmt.Sprintf("Unmapped item string ID %s", itemData.Name)}
	}
	blockRuntimeID = noBlockRuntimeID
	if blockStateData := itemData.Block; blockStateData != nil {
		id, ok := bedrock.RuntimeIDFor(blockStateData.Name, blockStateData.States)
		if !ok {
			panic("Unmapped blockstate returned by blockstate serializer: " + blockStateData.Name)
		}
		blockRuntimeID = id
	}
	return numericID, int16(itemData.Meta), blockRuntimeID, nil
}

// FromNetworkID is a port of ItemTranslator::fromNetworkId.
func (t *ItemTranslator) FromNetworkID(networkID int32, networkMeta int16, networkBlockRuntimeID int32) (item.Item, error) {
	stringID, ok := bedrock.ItemNameForRuntimeID(networkID)
	if !ok {
		return nil, &TypeConversionError{Message: fmt.Sprintf("Invalid network itemstack ID %d", networkID)}
	}

	var blockStateData *bedrock.BlockStateData
	if _, isBlockItem := t.blockItemIDs.LookupBlockID(stringID); isBlockItem {
		states := bedrock.BlockStates()
		if networkBlockRuntimeID < 0 || int(networkBlockRuntimeID) >= len(states) {
			return nil, &TypeConversionError{Message: fmt.Sprintf("Blockstate runtimeID %d does not correspond to any known blockstate", networkBlockRuntimeID)}
		}
		data := states[networkBlockRuntimeID]
		blockStateData = &data
	} else if networkBlockRuntimeID != noBlockRuntimeID {
		return nil, &TypeConversionError{Message: fmt.Sprintf("Item %s is not a blockitem, but runtime ID %d was provided", stringID, networkBlockRuntimeID)}
	}

	it, err := t.deserializer.DeserializeType(bedrockitem.SavedItemData{Name: stringID, Meta: int(networkMeta), Block: blockStateData})
	if err != nil {
		return nil, &TypeConversionError{Message: "Invalid network itemstack data: " + err.Error()}
	}
	return it, nil
}

// ItemTypeName is GlobalItemDataHandlers::getSerializer()->serializeType($item)->getName(): the
// item's saved type ID. ok is false for an item type with no serializer.
func ItemTypeName(it item.Item) (string, bool) {
	data, err := worldio.GetItemSerializer().SerializeType(it)
	if err != nil {
		return "", false
	}
	return data.Name, true
}

// DeserializeItemType is GlobalItemDataHandlers::getDeserializer()->deserializeType(new
// SavedItemData($name, $meta, $blockStateData)): ok is false for an unknown item type.
func DeserializeItemType(name string, meta int, blockStateData *bedrock.BlockStateData) (item.Item, bool) {
	it, err := worldio.GetItemDeserializer().DeserializeType(bedrockitem.SavedItemData{Name: name, Meta: meta, Block: blockStateData})
	if err != nil {
		return nil, false
	}
	return it, true
}

// ToNetworkNbt is a port of ItemTranslator::toNetworkNbt: the item's NBT as sent in tile spawn
// data (this relies on network item NBT being the same as disk item NBT, like PHP).
func (t *ItemTranslator) ToNetworkNbt(it item.Item) (*nbt.CompoundTag, error) {
	stack, err := t.serializer.SerializeStack(it, nil)
	if err != nil {
		return nil, err
	}
	return stack.ToNbt(), nil
}
