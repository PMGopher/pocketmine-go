// Package serializer is a port of a slice of pocketmine\network\mcpe\serializer\ChunkSerializer:
// turning a format.Chunk into the raw bytes a LevelChunk packet's RawPayload carries.
package serializer

import (
	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/format"
)

// Overworld subchunk index bounds (ChunkSerializer::getDimensionChunkBounds' DimensionIds::OVERWORLD
// case) - the only dimension this port sends chunks for so far. The Nether ([0,7]) and End
// ([0,15]) bounds aren't ported since nothing constructs a Nether/End World yet.
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

// SerializeFullChunk is a port of ChunkSerializer::serializeFullChunk, hard-coded to the overworld
// dimension and always using network (non-persistent, runtime-ID-based) block state IDs - the
// persistent (world-save, NBT-based) block state path isn't ported, since nothing in this port
// writes chunks to disk yet, only sends them over the network.
func SerializeFullChunk(chunk *format.Chunk, translator *convert.BlockTranslator) []byte {
	var buf []byte

	subChunkCount := GetSubChunkCount(chunk)
	writtenCount := 0
	for y := overworldMinSubChunkIndex; writtenCount < subChunkCount; y, writtenCount = y+1, writtenCount+1 {
		buf = append(buf, SerializeSubChunk(chunk.GetSubChunk(y), translator)...)
	}

	// "all biomes must always be written" - PHP's own comment on the loop below.
	for y := overworldMinSubChunkIndex; y <= overworldMaxSubChunkIndex; y++ {
		buf = append(buf, serializeBiomePalette(chunk.GetSubChunk(y).GetBiomeArray())...)
	}

	buf = append(buf, 0) // border block array count - always empty (see ChunkSerializer.php's own comment: these crash the regular client)

	// Tiles: this port has no Tile-in-World system yet (see format.Chunk's doc comment on why
	// tiles aren't part of Chunk at all here), so there's never anything to write - matching
	// ChunkSerializer::serializeTiles with an empty tile list.
	buf = append(buf, binaryutils.WriteUnsignedVarInt(0)...)

	return buf
}

// SerializeSubChunk is a port of ChunkSerializer::serializeSubChunk, always using network
// (non-persistent) block state IDs (the `$persistentBlockStates` parameter is always false here).
func SerializeSubChunk(subChunk *format.SubChunk, translator *convert.BlockTranslator) []byte {
	layers := subChunk.GetBlockLayers()
	buf := []byte{8, byte(len(layers))} // version, layer count

	for _, layer := range layers {
		bitsPerBlock := layer.GetBitsPerBlock()
		buf = append(buf, byte(bitsPerBlock<<1)|1) // |1 = non-persistent (network runtime IDs)
		buf = append(buf, layer.GetWordArray()...)

		palette := layer.GetPalette()
		if bitsPerBlock != 0 {
			buf = append(buf, binaryutils.WriteVarInt(int32(len(palette)))...)
		}
		for _, internalStateID := range palette {
			buf = append(buf, binaryutils.WriteVarInt(translator.NetworkIDForCachedState(internalStateID))...)
		}
	}
	return buf
}

// serializeBiomePalette is a port of ChunkSerializer::serializeBiomePalette. LegacyBiomeIdToStringIdMap
// isn't ported (no PocketMine-MP source checked out for it - like BiomeIds, it's part of the
// vendored pocketmine/bedrock-data-adjacent data this port hasn't needed to pull in yet), so this
// skips the "does this legacy biome ID have a valid string mapping" validation the PHP original
// does and writes every biome ID as-is. Every biome ID this port currently ever writes is a single
// hard-coded valid value (see world/generator's Flat generator), so that validation gap has no
// practical effect yet.
func serializeBiomePalette(biomes *format.PalettedBlockArray) []byte {
	bitsPerBlock := biomes.GetBitsPerBlock()
	buf := []byte{byte(bitsPerBlock<<1) | 1} // |1 = non-persistence bit; has no effect on biomes (always integer IDs), same as the PHP original's comment
	buf = append(buf, biomes.GetWordArray()...)

	palette := biomes.GetPalette()
	if bitsPerBlock != 0 {
		buf = append(buf, binaryutils.WriteVarInt(int32(len(palette)))...)
	}
	for _, biomeID := range palette {
		buf = append(buf, binaryutils.WriteVarInt(biomeID)...)
	}
	return buf
}
