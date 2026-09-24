package convert

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// sharedItemTranslator backs CoreItemStackToNet - ItemTranslator is stateless (see its doc
// comment), so one instance serves every caller.
var sharedItemTranslator = NewItemTranslator()

// CoreItemStackToNet is a port of TypeConverter::coreItemStackToNet: an item.Item as a network
// ItemStack, carrying its NBT.
//
// Not ported: cleanupUnnecessaryItemNBT's stripping of container-contents/block-entity NBT (only a
// bandwidth optimisation), and PHP's fallback of displaying unmapped items as INFO_UPDATE blocks -
// the info_update block isn't registered in this port yet, so an item with no network mapping is
// sent as an empty slot instead.
func CoreItemStackToNet(it item.Item) protocol.ItemStack {
	if it == nil || it.IsNull() {
		return protocol.ItemStack{}
	}
	networkID, meta, blockRuntimeID, ok := sharedItemTranslator.ToNetworkID(it)
	if !ok {
		return protocol.ItemStack{}
	}
	stack := protocol.ItemStack{
		ItemType:       protocol.ItemType{NetworkID: networkID, MetadataValue: uint32(meta)},
		BlockRuntimeID: blockRuntimeID,
		Count:          uint16(it.GetCount()),
		HasNetworkID:   true,
	}
	if tag := it.GetNamedTag(); tag.Count() > 0 {
		stack.NBTData = NbtToMap(tag)
	}
	return stack
}

// ItemStackWrapperLegacy is a port of ItemStackWrapper::legacy: a stack network ID of 0 for an
// empty slot and 1 otherwise (what PHP sends outside server-authoritative inventory requests).
func ItemStackWrapperLegacy(it item.Item) protocol.ItemInstance {
	stack := CoreItemStackToNet(it)
	stackNetworkID := int32(0)
	if stack.HasNetworkID {
		stackNetworkID = 1
	}
	return protocol.ItemInstance{StackNetworkID: stackNetworkID, Stack: stack}
}

// NbtToMap converts a pocketmine-go NBT compound into the map[string]any form gophertunnel encodes
// (TAG_Byte as uint8, TAG_Short as int16, and so on).
func NbtToMap(tag *nbt.CompoundTag) map[string]any {
	result := make(map[string]any, tag.Count())
	for name, value := range tag.All() {
		result[name] = nbtToAny(value)
	}
	return result
}

func nbtToAny(tag nbt.Tag) any {
	switch v := tag.(type) {
	case nbt.ByteTag:
		return uint8(v)
	case nbt.ShortTag:
		return int16(v)
	case nbt.IntTag:
		return int32(v)
	case nbt.LongTag:
		return int64(v)
	case nbt.FloatTag:
		return float32(v)
	case nbt.DoubleTag:
		return float64(v)
	case nbt.StringTag:
		return string(v)
	case nbt.ByteArrayTag:
		return []byte(v)
	case nbt.IntArrayTag:
		return []int32(v)
	case *nbt.CompoundTag:
		return NbtToMap(v)
	case *nbt.ListTag:
		values := v.Values()
		list := make([]any, len(values))
		for i, child := range values {
			list[i] = nbtToAny(child)
		}
		return list
	}
	return nil
}
