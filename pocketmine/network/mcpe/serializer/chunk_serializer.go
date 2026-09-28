// Package serializer is a port of a slice of pocketmine\network\mcpe\serializer\ChunkSerializer:
// turning a format.Chunk into the raw bytes a LevelChunk packet's RawPayload carries.
package serializer

import (
	"bytes"
	"math"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/format"
)

// Overworld subchunk index bounds (ChunkSerializer::getDimensionChunkBounds' DimensionIds::OVERWORLD
// case). PocketMine-MP sends every world as the overworld (NetworkSession always passes
// DimensionIds::OVERWORLD, a Nether-generated world included), so the other cases are never used.
const (
	overworldMinSubChunkIndex = format.MinSubChunkIndex
	overworldMaxSubChunkIndex = format.MaxSubChunkIndex
)

// GetSubChunkCount is a port of ChunkSerializer::getSubChunkCount, hard-coded to the overworld
// dimension bounds (see this file's doc comment) - chunks are sent as a stack, so every subchunk
// below the topmost non-empty one must be included even if some of those are themselves empty.
func GetSubChunkCount(chunk *format.Chunk) int {
	total := overworldMaxSubChunkIndex - overworldMinSubChunkIndex + 1
	for y, count := overworldMaxSubChunkIndex, total; y >= overworldMinSubChunkIndex; y, count = y-1, count-1 {
		if chunk.GetSubChunk(y).IsEmptyFast() {
			continue
		}
		return count
	}
	return 0
}

// SerializeFullChunk is a port of ChunkSerializer::serializeFullChunk for the overworld with
// network (runtime ID) block states, the only way PocketMine-MP calls it for the network; worlds
// are saved by the world providers' own serializers.
//
// cmd/pocketmine-go sends chunks in sub-chunk request mode instead (SerializeBiomesPayload +
// SerializeSubChunk), which is what Bedrock 1.26.50 servers known to work use; this full-chunk
// form is kept for the LevelChunk non-request path.
func SerializeFullChunk(chunk *format.Chunk, translator *convert.BlockTranslator) []byte {
	var buf []byte

	subChunkCount := GetSubChunkCount(chunk)
	writtenCount := 0
	for y := overworldMinSubChunkIndex; writtenCount < subChunkCount; y, writtenCount = y+1, writtenCount+1 {
		buf = append(buf, SerializeSubChunk(chunk.GetSubChunk(y), y, translator, nil)...)
	}

	buf = append(buf, SerializeBiomes(chunk)...)

	buf = append(buf, 0) // border block array count - always empty (see ChunkSerializer.php's own comment: these crash the regular client)

	buf = append(buf, SerializeTiles(chunk, AllSubChunks)...)

	return buf
}

// AllSubChunks makes SerializeTiles write the tiles of every sub-chunk.
const AllSubChunks = math.MinInt

// SerializeTiles is a port of ChunkSerializer::serializeTiles: the spawn compound (network NBT) of
// every Spawnable tile in the chunk, or only in sub-chunk subY (sub-chunk request mode).
func SerializeTiles(chunk *format.Chunk, subY int) []byte {
	var buf bytes.Buffer
	enc := gtnbt.NewEncoderWithEncoding(&buf, gtnbt.NetworkLittleEndian)
	for _, t := range chunk.GetTiles() {
		if subY != AllSubChunks && t.GetPosition().FloorY()>>4 != subY {
			continue
		}
		if compound, ok := tile.SerializedSpawnCompound(t); ok {
			_ = enc.Encode(convert.NbtToMap(compound))
		}
	}
	return buf.Bytes()
}

// SerializeBiomesPayload is the LevelChunk payload for sub-chunk request mode: every sub-chunk's
// biomes followed by the (always empty) border block count. The blocks themselves are sent later,
// one SubChunk packet entry per sub-chunk the client asks for.
func SerializeBiomesPayload(chunk *format.Chunk) []byte {
	return append(SerializeBiomes(chunk), 0)
}

// SerializeBiomes writes the biome palette of every overworld sub-chunk ("all biomes must always
// be written" - PHP's own comment on this loop in serializeFullChunk).
func SerializeBiomes(chunk *format.Chunk) []byte {
	var buf []byte
	for y := overworldMinSubChunkIndex; y <= overworldMaxSubChunkIndex; y++ {
		buf = append(buf, serializeBiomePalette(chunk.GetSubChunk(y).GetBiomeArray())...)
	}
	return buf
}

// SerializeSubChunk is a port of ChunkSerializer::serializeSubChunk, always using network
// (non-persistent) block state IDs (the `$persistentBlockStates` parameter is always false here).
//
// Unlike PocketMine-MP 5.44.4 (which only supports clients up to 1.26.30 and writes version 8),
// this writes sub-chunk format version 9, which carries the sub-chunk's absolute Y index after the
// layer count. Version 9 is what the vanilla server and Dragonfly send, and what the sub-chunk
// request system requires. y is the sub-chunk index (format.MinSubChunkIndex..MaxSubChunkIndex).
func SerializeSubChunk(subChunk *format.SubChunk, y int, translator *convert.BlockTranslator, blockAt func(x, y, z int) block.Behavior) []byte {
	layers := subChunk.GetBlockLayers()
	buf := []byte{9, byte(len(layers)), byte(int8(y))} // version, layer count, sub-chunk Y index

	for _, layer := range layers {
		networkIDs := false
		if blockAt != nil && hasNeighbourDependentStates(layer, translator) {
			layer = networkLayer(layer, translator, blockAt)
			networkIDs = true
		}
		bitsPerBlock := layer.GetBitsPerBlock()
		buf = append(buf, byte(bitsPerBlock<<1)|1) // |1 = non-persistent (network runtime IDs)
		buf = append(buf, layer.GetWordArray()...)

		palette := layer.GetPalette()
		if bitsPerBlock != 0 {
			buf = append(buf, binaryutils.WriteVarInt(int32(len(palette)))...)
		}
		for _, stateID := range palette {
			if !networkIDs {
				stateID = translator.NetworkIDForCachedState(stateID)
			}
			buf = append(buf, binaryutils.WriteVarInt(stateID)...)
		}
	}
	return buf
}

// hasNeighbourDependentStates reports whether layer has blocks whose 1.26.50 network state
// depends on their neighbours (fences, panes, bars, tripwire, stairs: see
// BlockTranslator.NetworkIDForBlock).
func hasNeighbourDependentStates(layer *format.PalettedBlockArray, translator *convert.BlockTranslator) bool {
	for _, stateID := range layer.GetPalette() {
		if translator.DependsOnNeighbours(int(stateID)) {
			return true
		}
	}
	return false
}

// networkLayer is layer with network runtime IDs instead of internal state IDs, the blocks that
// depend on their neighbours getting their state from the block read from the world (blockAt).
func networkLayer(layer *format.PalettedBlockArray, translator *convert.BlockTranslator, blockAt func(x, y, z int) block.Behavior) *format.PalettedBlockArray {
	var result *format.PalettedBlockArray
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := 0; y < format.SubChunkEdgeLength; y++ {
				stateID := layer.Get(x, y, z)
				var networkID int32
				if translator.DependsOnNeighbours(int(stateID)) {
					networkID = translator.NetworkIDForBlock(blockAt(x, y, z))
				} else {
					networkID = translator.NetworkIDForCachedState(stateID)
				}
				if result == nil {
					result = format.NewPalettedBlockArray(networkID)
				}
				result.Set(x, y, z, networkID)
			}
		}
	}
	return result
}

// serializeBiomePalette is a port of ChunkSerializer::serializeBiomePalette: biome IDs the client
// doesn't know are sent as ocean (the 1.18.0 client crashes on bogus biomes, PHP's comment says).
func serializeBiomePalette(biomes *format.PalettedBlockArray) []byte {
	bitsPerBlock := biomes.GetBitsPerBlock()
	buf := []byte{byte(bitsPerBlock<<1) | 1} // |1 = non-persistence bit; has no effect on biomes (always integer IDs), same as the PHP original's comment
	buf = append(buf, biomes.GetWordArray()...)

	palette := biomes.GetPalette()
	if bitsPerBlock != 0 {
		buf = append(buf, binaryutils.WriteVarInt(int32(len(palette)))...)
	}
	biomeIDMap := bedrock.GetLegacyBiomeIdToStringIdMap()
	for _, biomeID := range palette {
		if _, known := biomeIDMap.LegacyToString(int(biomeID)); !known {
			biomeID = 0 // BiomeIds::OCEAN
		}
		buf = append(buf, binaryutils.WriteVarInt(biomeID)...)
	}
	return buf
}
