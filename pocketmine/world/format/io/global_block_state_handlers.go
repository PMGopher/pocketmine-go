package io

import (
	"sync"

	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
)

// GlobalBlockStateHandlers is a port of pocketmine\world\format\io\GlobalBlockStateHandlers: the
// shared block state serializer and deserializer (VanillaBlockMappings), used for world saves and,
// through BlockTranslator, the network. The block data upgrader (loading states saved by older
// versions) isn't ported.
var (
	globalRegistrar     *blockconvert.BlockSerializerDeserializerRegistrar
	globalRegistrarOnce sync.Once
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

// GetUnknownBlockStateData is GlobalBlockStateHandlers::getUnknownBlockStateData.
func GetUnknownBlockStateData() bedrock.BlockStateData {
	return bedrock.BlockStateData{Name: ids.INFO_UPDATE, States: map[string]any{}, Version: blockconvert.CurrentBlockStateVersion}
}
