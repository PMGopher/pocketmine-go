package main

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/format"
)

// Chunks are sent using the sub-chunk request system (what the vanilla server and Dragonfly use for
// Bedrock 1.26.50): LevelChunk only carries the biomes and how many sub-chunks exist, and the client
// then asks for the sub-chunks it needs with SubChunkRequest, answered by handleSubChunkRequest.
// PocketMine-MP 5.44.4 still sends whole chunks (sub-chunk format version 8) in LevelChunk, but it
// only supports clients up to 1.26.30, so there is no PHP code to port for 1.26.50 here.

// levelChunkPacket builds the request-mode LevelChunk packet for the chunk at chunkX/chunkZ.
func levelChunkPacket(chunkX, chunkZ int, chunk *format.Chunk) *packet.LevelChunk {
	return &packet.LevelChunk{
		Position:      protocol.ChunkPos{int32(chunkX), int32(chunkZ)},
		SubChunkCount: 0,
		SubChunkLimit: protocol.Option(int32(serializer.GetSubChunkCount(chunk))),
		RawPayload:    serializer.SerializeBiomesPayload(chunk),
	}
}

// handleSubChunkRequest answers a SubChunkRequest with one entry per requested offset.
func handleSubChunkRequest(w *world.World, pk *packet.SubChunkRequest) *packet.SubChunk {
	entries := make([]protocol.SubChunkEntry, 0, len(pk.Offsets))
	for _, offset := range pk.Offsets {
		entries = append(entries, subChunkEntry(w, pk.Position, offset))
	}
	return &packet.SubChunk{
		Dimension:       pk.Dimension,
		Position:        pk.Position,
		SubChunkEntries: entries,
	}
}

// subChunkEntry serialises the sub-chunk at centre+offset (centre's Y is an absolute sub-chunk
// index, format.MinSubChunkIndex..MaxSubChunkIndex for the overworld).
func subChunkEntry(w *world.World, centre protocol.SubChunkPos, offset protocol.SubChunkOffset) protocol.SubChunkEntry {
	subY := int(centre[1]) + int(offset[1])
	if subY < format.MinSubChunkIndex || subY > format.MaxSubChunkIndex {
		return protocol.SubChunkEntry{Result: protocol.SubChunkResultIndexOutOfBounds, Offset: offset}
	}
	chunk, ok := w.GetChunk(int(centre[0])+int(offset[0]), int(centre[2])+int(offset[2]))
	if !ok {
		return protocol.SubChunkEntry{Result: protocol.SubChunkResultChunkNotFound, Offset: offset}
	}

	heightMapType, heightMap := subChunkHeightMap(chunk, subY)
	entry := protocol.SubChunkEntry{
		Offset:              offset,
		HeightMapType:       heightMapType,
		HeightMapData:       heightMap,
		RenderHeightMapType: heightMapType,
		RenderHeightMapData: heightMap,
	}

	sub := chunk.GetSubChunk(subY)
	if sub.IsEmptyFast() {
		entry.Result = protocol.SubChunkResultSuccessAllAir
		return entry
	}
	entry.Result = protocol.SubChunkResultSuccess
	// Tiles would follow the sub-chunk data; this port has no tiles in chunks yet (see
	// serializer.SerializeFullChunk).
	entry.RawPayload = protocol.Option(serializer.SerializeSubChunk(sub, subY, w.Translator()))
	return entry
}

// subChunkHeightMap builds the per-column height map of sub-chunk subY: for every column, the height
// of its highest block relative to the sub-chunk's base, 16 if it's in a sub-chunk above and -1 if
// it's below. If every column is above (or below), only the type is sent.
func subChunkHeightMap(chunk *format.Chunk, subY int) (byte, protocol.Optional[protocol.HeightMap]) {
	var heightMap protocol.HeightMap
	allHigher, allLower := true, true
	base := subY << 4
	for z := 0; z < 16; z++ {
		for x := 0; x < 16; x++ {
			y, ok := chunk.GetHighestBlockAt(x, z)
			if !ok {
				y = format.MinSubChunkIndex << 4
			}
			switch other := y >> 4; {
			case other > subY:
				heightMap[z][x], allLower = 16, false
			case other < subY:
				heightMap[z][x], allHigher = -1, false
			default:
				heightMap[z][x], allLower, allHigher = int8(y-base), false, false
			}
		}
	}
	switch {
	case allHigher:
		return protocol.HeightMapDataTooHigh, protocol.Optional[protocol.HeightMap]{}
	case allLower:
		return protocol.HeightMapDataTooLow, protocol.Optional[protocol.HeightMap]{}
	}
	return protocol.HeightMapDataHasData, protocol.Option(heightMap)
}
