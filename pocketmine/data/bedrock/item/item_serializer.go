package bedrockitem

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	"pocketmine-go/pocketmine/item"
)

// ItemSerializer is a port of pocketmine\data\bedrock\item\ItemSerializer.
type ItemSerializer struct {
	blockStateSerializer *blockconvert.BlockObjectToStateSerializer

	itemSerializers      map[int]func(it item.Item) SavedItemData
	blockItemSerializers map[int]func(blk block.Behavior) SavedItemData
}

func NewItemSerializer(blockStateSerializer *blockconvert.BlockObjectToStateSerializer) *ItemSerializer {
	s := &ItemSerializer{
		blockStateSerializer: blockStateSerializer,
		itemSerializers:      map[int]func(item.Item) SavedItemData{},
		blockItemSerializers: map[int]func(block.Behavior) SavedItemData{},
	}
	s.registerSpecialBlockSerializers()
	NewItemSerializerDeserializerRegistrar(nil, s)
	return s
}

// Map is a port of ItemSerializer::map.
func (s *ItemSerializer) Map(it item.Item, serializer func(it item.Item) SavedItemData) {
	index := it.GetTypeId()
	if _, ok := s.itemSerializers[index]; ok {
		panic(fmt.Sprintf("Item type ID %d already has a serializer registered", index))
	}
	s.itemSerializers[index] = serializer
}

// MapBlock is a port of ItemSerializer::mapBlock.
func (s *ItemSerializer) MapBlock(blk block.Behavior, serializer func(blk block.Behavior) SavedItemData) {
	index := blk.GetTypeId()
	if _, ok := s.blockItemSerializers[index]; ok {
		panic(fmt.Sprintf("Block type ID %d already has a serializer registered", index))
	}
	s.blockItemSerializers[index] = serializer
}

// SerializeType is a port of ItemSerializer::serializeType.
func (s *ItemSerializer) SerializeType(it item.Item) (data SavedItemData, err error) {
	defer recoverItemError(&err)
	if it.IsNull() {
		panic("Cannot serialize a null itemstack")
	}
	if itemBlock, ok := it.(*item.ItemBlock); ok {
		data = s.serializeBlockItem(itemBlock.GetBlock())
	} else {
		index := it.GetTypeId()
		serializer, ok := s.itemSerializers[index]
		if !ok {
			return SavedItemData{}, &ItemTypeSerializeError{Message: fmt.Sprintf("No serializer registered for %T (%d) %s", it, index, it.GetName())}
		}
		data = serializer(it)
	}

	resultTag := it.GetNamedTag()
	if resultTag.Count() > 0 {
		if extraTag := data.Tag; extraTag != nil {
			resultTag = resultTag.Merge(extraTag)
		}
		data = SavedItemData{Name: data.Name, Meta: data.Meta, Block: data.Block, Tag: resultTag}
	}
	return data, nil
}

func (s *ItemSerializer) serializeBlockItem(blk block.Behavior) SavedItemData {
	if serializer, ok := s.blockItemSerializers[blk.GetTypeId()]; ok {
		return serializer(blk)
	}
	return s.standardBlock(blk)
}

// standardBlock is a port of ItemSerializer::standardBlock.
func (s *ItemSerializer) standardBlock(blk block.Behavior) SavedItemData {
	blockStateData, err := s.blockStateSerializer.Serialize(blk.GetStateId())
	if err != nil {
		panic(&ItemTypeSerializeError{Message: err.Error()})
	}
	//TODO: this really ought to throw if there's no blockitem ID
	itemNameID, ok := GetBlockItemIdMap().LookupItemID(blockStateData.Name)
	if !ok {
		itemNameID = blockStateData.Name
	}
	return SavedItemData{Name: itemNameID, Block: &blockStateData}
}

// registerSpecialBlockSerializers is a port of ItemSerializer::registerSpecialBlockSerializers.
func (s *ItemSerializer) registerSpecialBlockSerializers() {
	//these are encoded as regular blocks, but they have to be accounted for explicitly since they don't use ItemBlock
	//Bamboo->getBlock() returns BambooSapling :(
	s.Map(item.VanillaItem("bamboo"), func(item.Item) SavedItemData { return s.standardBlock(block.VanillaBlock("bamboo")) })
	s.Map(item.VanillaItem("coral_fan"), func(it item.Item) SavedItemData { return s.standardBlock(it.GetBlock()) })
}
