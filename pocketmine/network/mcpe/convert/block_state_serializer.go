// Package convert is a port of a small slice of pocketmine\network\mcpe\convert and
// pocketmine\data\bedrock\block\convert: translating this port's own internal block state IDs
// (block.Behavior.GetStateId(), already used throughout the block package's DescribeBlockOnlyState/
// DescribeBlockItemState machinery) into the Bedrock network blockstate NBT data connecting
// clients actually expect.
//
// The real PocketMine-MP's equivalent (VanillaBlockMappings.php) registers a serializer for every
// one of ~700 vanilla blocks in ~1700 lines - a whole separate undertaking from VanillaBlocks
// itself, on the same scale. This starts with just the handful of block types the Flat world
// generator needs (see pocketmine/world/generator), registered the same "port what's needed,
// extend incrementally" way VanillaBlocks itself started - not a shortcut standing in for the real
// architecture, just an early, small slice of it.
package convert

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockStateSerializerFunc is a port of one blockstate-serializing closure as registered via
// BlockSerializerDeserializerRegistrar in the PHP original (VanillaBlockMappings::setupBlockStateSerializer,
// mapSimple/mapModel/... ultimately all resolve to a closure of this shape).
type BlockStateSerializerFunc func(blk block.Behavior) (bedrock.BlockStateData, error)

var blockStateSerializers = map[int]BlockStateSerializerFunc{}

// RegisterBlockStateSerializer registers the serializer for one block type ID. Panics on a
// duplicate registration - a programmer error at the call site, the same convention used for
// BlockIdentifier's own uniqueness checks elsewhere in this port.
func RegisterBlockStateSerializer(blockTypeID int, fn BlockStateSerializerFunc) {
	if _, exists := blockStateSerializers[blockTypeID]; exists {
		panic(fmt.Sprintf("convert: duplicate BlockStateSerializer registration for block type ID %d", blockTypeID))
	}
	blockStateSerializers[blockTypeID] = fn
}

// SerializeBlockState is a port of BlockObjectToStateSerializer::serialize (the entry point
// BlockStateSerializer::serialize($internalStateId) uses internally, adapted to take the already-
// decoded Behavior directly rather than re-deriving it from a bare internal state ID - this port's
// World doesn't have a "construct a Behavior from a bare internal state ID" lookup yet, so callers
// that have a live Behavior in hand (every caller today) pass it directly).
func SerializeBlockState(blk block.Behavior) (bedrock.BlockStateData, error) {
	fn, ok := blockStateSerializers[blk.GetTypeId()]
	if !ok {
		return bedrock.BlockStateData{}, fmt.Errorf("convert: no BlockStateSerializer registered for block type ID %d (%s)", blk.GetTypeId(), blk.GetName())
	}
	return fn(blk)
}
