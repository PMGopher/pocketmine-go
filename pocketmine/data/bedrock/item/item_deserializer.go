package bedrockitem

import (
	"errors"
	"fmt"

	"pocketmine-go/pocketmine/block"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	"pocketmine-go/pocketmine/item"
)

// ItemDeserializer is a port of pocketmine\data\bedrock\item\ItemDeserializer.
type ItemDeserializer struct {
	blockStateDeserializer *blockconvert.BlockStateToObjectDeserializer

	deserializers map[string]func(data SavedItemData) item.Item
}

func NewItemDeserializer(blockStateDeserializer *blockconvert.BlockStateToObjectDeserializer) *ItemDeserializer {
	d := &ItemDeserializer{blockStateDeserializer: blockStateDeserializer, deserializers: map[string]func(SavedItemData) item.Item{}}
	NewItemSerializerDeserializerRegistrar(d, nil)
	return d
}

// Map is a port of ItemDeserializer::map.
func (d *ItemDeserializer) Map(id string, deserializer func(data SavedItemData) item.Item) {
	d.deserializers[id] = deserializer
}

// GetDeserializerForID is a port of ItemDeserializer::getDeserializerForId.
func (d *ItemDeserializer) GetDeserializerForID(id string) (func(data SavedItemData) item.Item, bool) {
	f, ok := d.deserializers[id]
	return f, ok
}

// MapBlock is a port of ItemDeserializer::mapBlock.
func (d *ItemDeserializer) MapBlock(id string, deserializer func(data SavedItemData) block.Behavior) {
	d.Map(id, func(data SavedItemData) item.Item { return blockAsItem(deserializer(data)) })
}

// blockAsItem is Block::asItem.
func blockAsItem(blk block.Behavior) item.Item {
	it, err := blk.(interface{ AsItem() (block.Item, error) }).AsItem()
	if err != nil {
		panic(&ItemTypeDeserializeError{Message: err.Error()})
	}
	return it.(item.Item)
}

// DeserializeType is a port of ItemDeserializer::deserializeType.
func (d *ItemDeserializer) DeserializeType(data SavedItemData) (it item.Item, err error) {
	defer recoverItemError(&err)
	if blockData := data.Block; blockData != nil {
		//TODO: this is rough duct tape; we need a better way to deal with this
		stateID, err := d.blockStateDeserializer.Deserialize(*blockData)
		if err != nil {
			var unsupported *blockconvert.UnsupportedBlockStateError
			if errors.As(err, &unsupported) {
				return nil, &UnsupportedItemTypeError{Message: err.Error()}
			}
			return nil, &ItemTypeDeserializeError{Message: "Failed to deserialize item data: " + err.Error()}
		}
		//TODO: worth caching this or not?
		return blockAsItem(block.GetRuntimeBlockStateRegistry().FromStateId(stateID)), nil
	}
	id := data.Name
	deserializer, ok := d.deserializers[id]
	if !ok {
		return nil, &UnsupportedItemTypeError{Message: fmt.Sprintf("No deserializer found for ID %s", id)}
	}
	return deserializer(data), nil
}

// DeserializeStack is a port of ItemDeserializer::deserializeStack.
func (d *ItemDeserializer) DeserializeStack(data SavedItemStackData) (it item.Item, err error) {
	itemStack, err := d.DeserializeType(data.TypeData)
	if err != nil {
		return nil, err
	}
	itemStack.SetCount(data.Count)
	if tag := data.TypeData.Tag; tag != nil {
		// The item's setters panic on invalid values, like PHP's NbtException.
		func() {
			defer func() {
				if p := recover(); p != nil {
					err = &ItemTypeDeserializeError{Message: fmt.Sprintf("Invalid item saved NBT: %v", p)}
				}
			}()
			itemStack.SetNamedTag(tag.Clone())
		}()
		if err != nil {
			return nil, err
		}
	}

	//TODO: this hack is necessary to get legacy tools working - we need a better way to handle this kind of stuff
	if durable, ok := itemStack.(interface {
		GetDamage() int
		SetDamage(int)
		GetMaxDurability() int
	}); ok && durable.GetDamage() == 0 && data.TypeData.Meta > 0 {
		durable.SetDamage(min(data.TypeData.Meta, durable.GetMaxDurability()))
	}

	//TODO: canDestroy, canPlaceOn, wasPickedUp are currently unused
	return itemStack, nil
}
