package io

import (
	"sync"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/nbt"

	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	blockupgrade "pocketmine-go/pocketmine/data/bedrock/block/upgrade"
)

// GlobalBlockStateHandlers is a port of pocketmine\world\format\io\GlobalBlockStateHandlers: the
// shared block state serializer and deserializer (VanillaBlockMappings), used for world saves and,
// through BlockTranslator, the network, and the block data upgrader for states saved by older
// versions.
var (
	globalRegistrar     *blockconvert.BlockSerializerDeserializerRegistrar
	globalRegistrarOnce sync.Once

	blockDataUpgrader     *blockupgrade.BlockDataUpgrader
	blockDataUpgraderOnce sync.Once
)

// GetBlockStateRegistrar is GlobalBlockStateHandlers::getRegistrar.
func GetBlockStateRegistrar() *blockconvert.BlockSerializerDeserializerRegistrar {
	globalRegistrarOnce.Do(func() {
		deserializer := blockconvert.NewBlockStateToObjectDeserializer()
		serializer := blockconvert.NewBlockObjectToStateSerializer()
		globalRegistrar = blockconvert.NewBlockSerializerDeserializerRegistrar(deserializer, serializer)
		blockconvert.InitVanillaBlockMappings(globalRegistrar)
	})
	return globalRegistrar
}

// GetBlockStateDeserializer is GlobalBlockStateHandlers::getDeserializer.
func GetBlockStateDeserializer() *blockconvert.BlockStateToObjectDeserializer {
	return GetBlockStateRegistrar().Deserializer
}

// GetBlockStateSerializer is GlobalBlockStateHandlers::getSerializer.
func GetBlockStateSerializer() *blockconvert.BlockObjectToStateSerializer {
	return GetBlockStateRegistrar().Serializer
}

// GetBlockDataUpgrader is GlobalBlockStateHandlers::getUpgrader: every nbt_upgrade_schema file
// and the 1.12.0 id_meta_to_nbt table (vendored in data/bedrock/block/upgrade/schema).
func GetBlockDataUpgrader() *blockupgrade.BlockDataUpgrader {
	blockDataUpgraderOnce.Do(func() {
		u, err := blockupgrade.NewDefaultBlockDataUpgrader()
		if err != nil {
			// The data is embedded; this can only fail if it's corrupt.
			panic("loading block upgrade schemas: " + err.Error())
		}
		blockDataUpgrader = u
	})
	return blockDataUpgrader
}

// GetUnknownBlockStateData is GlobalBlockStateHandlers::getUnknownBlockStateData.
func GetUnknownBlockStateData() bedrock.BlockStateData {
	return bedrock.NewCurrentBlockStateData(ids.INFO_UPDATE, nil)
}

// init gives the tile package FlowerPot's block state (de)serialization, which goes through these
// handlers (tile can't import this package).
func init() {
	tile.LegacyEntityIdToStringFunc = func(legacy int) (string, bool) {
		return bedrock.LegacyEntityIdToStringIdMap().LegacyToString(legacy)
	}
	fromData := func(data bedrock.BlockStateData) (tile.Block, error) {
		stateID, err := GetBlockStateDeserializer().Deserialize(data)
		if err != nil {
			return nil, err
		}
		return block.GetRuntimeBlockStateRegistry().FromStateId(stateID), nil
	}
	tile.BlockFromLegacyIdMetaFunc = func(id, meta int) (tile.Block, error) {
		data, err := GetBlockDataUpgrader().UpgradeIntIdMeta(id, meta)
		if err != nil {
			return nil, err
		}
		return fromData(data)
	}
	tile.BlockFromNbtFunc = func(tag *nbt.CompoundTag) (tile.Block, error) {
		data, err := GetBlockDataUpgrader().UpgradeBlockStateNbt(tag)
		if err != nil {
			return nil, err
		}
		return fromData(data)
	}
	tile.BlockToNbtFunc = func(b tile.Block) (*nbt.CompoundTag, error) {
		data, err := GetBlockStateSerializer().Serialize(b.GetStateId())
		if err != nil {
			return nil, err
		}
		return data.ToNbt(), nil
	}
	tile.IsAirFunc = func(b tile.Block) bool {
		_, isAir := b.(*block.Air)
		return isAir
	}
}
