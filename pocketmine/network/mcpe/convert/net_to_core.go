package convert

import (
	"fmt"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// TypeConversionError is a port of pocketmine\network\mcpe\convert\TypeConversionException.
type TypeConversionError struct{ Message string }

func (e *TypeConversionError) Error() string { return e.Message }

var (
	itemTypeByNameOnce sync.Once
	itemTypeByName     map[string]int
)

// DeserializeItemType is ItemDeserializer::deserializeType for a saved item type name and meta:
// ok is false for a type (or meta variant) that can't be deserialized yet (only the item types
// itemTypeNames maps, without meta variants or block items - see ItemTranslator).
func DeserializeItemType(name string, meta int) (item.Item, bool) {
	itemTypeByNameOnce.Do(func() {
		itemTypeByName = make(map[string]int, len(itemTypeNames))
		for typeID, n := range itemTypeNames {
			itemTypeByName[n] = typeID
		}
	})
	typeID, ok := itemTypeByName[name]
	if !ok || meta != 0 {
		return nil, false
	}
	constructor, ok := coreItemConstructors[typeID]
	if !ok {
		return nil, false
	}
	return constructor(), true
}

// FromNetworkID is a port of ItemTranslator::fromNetworkId: the item for a network item ID and
// meta. Only the item types itemTypeNames maps can be deserialized so far (no block items, no
// meta variants - see ItemTranslator).
func (t *ItemTranslator) FromNetworkID(networkID int32, meta int16, blockRuntimeID int32) (item.Item, error) {
	name, ok := bedrock.ItemNameForRuntimeID(networkID)
	if !ok {
		return nil, &TypeConversionError{Message: fmt.Sprintf("Unknown network item ID %d", networkID)}
	}
	if blockRuntimeID != noBlockRuntimeID {
		return nil, &TypeConversionError{Message: fmt.Sprintf("Unsupported network item %s (meta %d, block runtime ID %d)", name, meta, blockRuntimeID)}
	}
	it, ok := DeserializeItemType(name, int(meta))
	if !ok {
		return nil, &TypeConversionError{Message: fmt.Sprintf("Unsupported network item %s (meta %d, block runtime ID %d)", name, meta, blockRuntimeID)}
	}
	return it, nil
}

// NetItemStackToCore is a port of TypeConverter::netItemStackToCore.
func NetItemStackToCore(stack protocol.ItemStack) (item.Item, error) {
	if stack.NetworkID == 0 {
		return item.VanillaAir(), nil
	}
	result, err := sharedItemTranslator.FromNetworkID(stack.NetworkID, int16(stack.MetadataValue), stack.BlockRuntimeID)
	if err != nil {
		return nil, err
	}
	result.SetCount(int(stack.Count))
	if len(stack.NBTData) > 0 {
		result.SetNamedTag(MapToNbt(stack.NBTData))
	}
	return result, nil
}

// MapToNbt is the reverse of NbtToMap: gophertunnel's map form of an NBT compound as a
// pocketmine-go CompoundTag.
func MapToNbt(m map[string]any) *nbt.CompoundTag {
	tag := nbt.NewCompoundTag()
	for name, value := range m {
		if t := anyToNbt(value); t != nil {
			tag.SetTag(name, t)
		}
	}
	return tag
}

func anyToNbt(value any) nbt.Tag {
	switch v := value.(type) {
	case uint8:
		return nbt.ByteTag(int8(v))
	case int8:
		return nbt.ByteTag(v)
	case bool:
		if v {
			return nbt.ByteTag(1)
		}
		return nbt.ByteTag(0)
	case int16:
		return nbt.ShortTag(v)
	case int32:
		return nbt.IntTag(v)
	case int64:
		return nbt.LongTag(v)
	case float32:
		return nbt.FloatTag(v)
	case float64:
		return nbt.DoubleTag(v)
	case string:
		return nbt.StringTag(v)
	case []byte:
		return nbt.ByteArrayTag(v)
	case []int32:
		return nbt.IntArrayTag(v)
	case map[string]any:
		return MapToNbt(v)
	case []any:
		var tags []nbt.Tag
		for _, child := range v {
			if t := anyToNbt(child); t != nil {
				tags = append(tags, t)
			}
		}
		return newListTag(tags)
	case []map[string]any:
		var tags []nbt.Tag
		for _, child := range v {
			tags = append(tags, MapToNbt(child))
		}
		return newListTag(tags)
	}
	return nil
}

// Protocol game modes, ProtocolGameMode::*.
const (
	protocolGameModeSurvival       = 0
	protocolGameModeCreative       = 1
	protocolGameModeAdventure      = 2
	protocolGameModeSurvivalViewer = 3
	protocolGameModeCreativeViewer = 4
)

// ProtocolGameModeToCore is a port of TypeConverter::protocolGameModeToCore: the core game mode
// (player.GameMode value) for a protocol game mode, or false if it has none.
func ProtocolGameModeToCore(gameMode int32) (int, bool) {
	switch gameMode {
	case protocolGameModeSurvival:
		return coreGameModeSurvival, true
	case protocolGameModeCreative:
		return coreGameModeCreative, true
	case protocolGameModeAdventure:
		return coreGameModeAdventure, true
	case protocolGameModeSurvivalViewer, protocolGameModeCreativeViewer:
		//TODO: native spectator support
		return coreGameModeSpectator, true
	}
	return 0, false
}

// newListTag builds a list from decoded values (all of one type, which gophertunnel guarantees
// for data it decoded), or nil if they're mixed.
func newListTag(tags []nbt.Tag) nbt.Tag {
	list, err := nbt.NewListTag(tags, nbt.TagEnd)
	if err != nil {
		return nil
	}
	return list
}
