package mcpe

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld() *world.World {
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
	})
}

func TestLevelChunkPacketUsesRequestMode(t *testing.T) {
	w := newTestWorld()
	chunk := w.GetOrLoadChunk(2, -3)
	pk := LevelChunkPacket(2, -3, chunk, nil)
	limit, ok := pk.SubChunkLimit.Value()
	if pk.SubChunkCount != 0 || !ok || limit != int32(serializer.GetSubChunkCount(chunk)) {
		t.Errorf("LevelChunk count=%d limit=%d(%v), want 0 and the sub-chunk count", pk.SubChunkCount, limit, ok)
	}
	if pk.Position != (protocol.ChunkPos{2, -3}) {
		t.Errorf("position = %v", pk.Position)
	}
}

func TestHandleSubChunkRequest(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	// The vanilla flat world's surface is at y=63 (sub-chunk 3). Request the surface sub-chunk, the
	// one above it, one buried under it, one below the bottom of the world and one in a chunk that
	// isn't loaded.
	resp := HandleSubChunkRequest(w, &packet.SubChunkRequest{
		Position: protocol.SubChunkPos{0, 3, 0},
		Offsets:  []protocol.SubChunkOffset{{0, 0, 0}, {0, 1, 0}, {0, -3, 0}, {0, -8, 0}, {5, 0, 5}},
	}, nil)
	if len(resp.SubChunkEntries) != 5 {
		t.Fatalf("got %d entries, want 5", len(resp.SubChunkEntries))
	}
	terrain, air, buried, below, missing := resp.SubChunkEntries[0], resp.SubChunkEntries[1], resp.SubChunkEntries[2], resp.SubChunkEntries[3], resp.SubChunkEntries[4]

	if terrain.Result != protocol.SubChunkResultSuccess {
		t.Errorf("terrain result = %d, want success", terrain.Result)
	}
	if payload, ok := terrain.RawPayload.Value(); !ok || len(payload) < 3 || payload[0] != 9 || int8(payload[2]) != 3 {
		t.Errorf("terrain payload = %v, want a version 9 sub-chunk at y index 3", payload)
	}
	if h, ok := terrain.HeightMapData.Value(); !ok || h[0][0] != 15 {
		t.Errorf("terrain heightmap = %v, want the surface at 15 within the sub-chunk", h[0][0])
	}
	if terrain.HeightMapType != protocol.HeightMapDataHasData {
		t.Errorf("terrain heightmap type = %d, want HasData", terrain.HeightMapType)
	}
	if air.Result != protocol.SubChunkResultSuccessAllAir || air.HeightMapType != protocol.HeightMapDataTooLow {
		t.Errorf("air above: result=%d heightmap=%d, want all-air and too-low", air.Result, air.HeightMapType)
	}
	if buried.Result != protocol.SubChunkResultSuccess || buried.HeightMapType != protocol.HeightMapDataTooHigh {
		t.Errorf("buried: result=%d heightmap=%d, want success and too-high", buried.Result, buried.HeightMapType)
	}
	if below.Result != protocol.SubChunkResultIndexOutOfBounds {
		t.Errorf("below the world: result=%d, want index out of bounds", below.Result)
	}
	if missing.Result != protocol.SubChunkResultChunkNotFound {
		t.Errorf("unloaded chunk: result=%d, want chunk not found", missing.Result)
	}
}

func TestChunksThroughTheClientBlobCache(t *testing.T) {
	w := newTestWorld()
	chunk := w.GetOrLoadChunk(0, 0)
	cache := NewClientBlobCache()

	lc := LevelChunkPacket(0, 0, chunk, cache)
	if !lc.CacheEnabled || len(lc.BlobHashes) != 1 || len(lc.RawPayload) != 1 {
		t.Fatalf("LevelChunk cache=%v hashes=%d payload=%d, want the biomes as one blob and a 1-byte payload", lc.CacheEnabled, len(lc.BlobHashes), len(lc.RawPayload))
	}

	resp := HandleSubChunkRequest(w, &packet.SubChunkRequest{Position: protocol.SubChunkPos{0, 3, 0}, Offsets: []protocol.SubChunkOffset{{0, 0, 0}}}, cache)
	entry := resp.SubChunkEntries[0]
	hash, ok := entry.BlobHash.Value()
	if !resp.CacheEnabled || !ok {
		t.Fatalf("SubChunk cache=%v blob hash present=%v, want both", resp.CacheEnabled, ok)
	}
	if payload, _ := entry.RawPayload.Value(); len(payload) != 0 {
		t.Errorf("cached entry payload = %d bytes, want only the (empty) tiles", len(payload))
	}

	miss := cache.HandleBlobStatus(&packet.ClientCacheBlobStatus{MissHashes: []uint64{hash, lc.BlobHashes[0]}})
	if miss == nil || len(miss.Blobs) != 2 {
		t.Fatalf("miss response = %v, want both blobs", miss)
	}
	if miss.Blobs[0].Payload[0] != 9 {
		t.Errorf("sub-chunk blob starts with %d, want sub-chunk version 9", miss.Blobs[0].Payload[0])
	}
	if again := cache.HandleBlobStatus(&packet.ClientCacheBlobStatus{MissHashes: []uint64{hash}}); again != nil {
		t.Error("a blob that was already sent was sent again")
	}
}
