package leveldb

import (
	"errors"
	"testing"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

func newProvider(t *testing.T) *LevelDB {
	t.Helper()
	dir := t.TempDir()
	if err := Generate(dir, "test", io.WorldCreationOptions{GeneratorName: "normal"}); err != nil {
		t.Fatal(err)
	}
	p, err := NewLevelDB(dir, log.NewSimpleLogger())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p
}

func stateName(t *testing.T, stateID int32) string {
	t.Helper()
	return block.GetRuntimeBlockStateRegistry().FromStateId(int(stateID)).GetName()
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := newProvider(t)
	stone := int32(block.VanillaStone().GetStateId())
	sc := format.NewSubChunk(io.EmptyStateID(), nil, format.NewPalettedBlockArray(1))
	sc.SetBlockStateID(1, 2, 3, stone)
	subChunks := map[int]*format.SubChunk{}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		subChunks[y] = format.NewSubChunk(io.EmptyStateID(), nil, format.NewPalettedBlockArray(1))
	}
	subChunks[2] = sc
	tileTag := nbt.NewCompoundTag().SetString("id", "Chest").SetInt("x", 1).SetInt("y", 34).SetInt("z", 3)
	entityTag := nbt.NewCompoundTag().SetString("identifier", "minecraft:cow")
	if err := p.SaveChunk(5, -7, io.NewChunkData(subChunks, true, []*nbt.CompoundTag{entityTag}, []*nbt.CompoundTag{tileTag}), format.DirtyFlagsAll); err != nil {
		t.Fatal(err)
	}

	loaded, err := p.LoadChunk(5, -7)
	if err != nil || loaded == nil {
		t.Fatalf("LoadChunk: %v %v", loaded, err)
	}
	if loaded.IsUpgraded() {
		t.Errorf("a current-format chunk shouldn't be marked as upgraded")
	}
	data := loaded.GetData()
	if !data.IsPopulated() {
		t.Errorf("populated flag lost")
	}
	if got := data.GetSubChunks()[2].GetBlockStateID(1, 2, 3); got != stone {
		t.Errorf("block = %s, want stone", stateName(t, got))
	}
	if got := data.GetSubChunks()[2].GetBiomeArray().Get(0, 0, 0); got != 1 {
		t.Errorf("biome = %d, want 1", got)
	}
	if len(data.GetTileNBT()) != 1 || len(data.GetEntityNBT()) != 1 {
		t.Errorf("tiles/entities: %d/%d", len(data.GetTileNBT()), len(data.GetEntityNBT()))
	}
	if missing, err := p.LoadChunk(100, 100); missing != nil || err != nil {
		t.Errorf("a missing chunk should be nil, nil: %v %v", missing, err)
	}
	if n, _ := p.CalculateChunkCount(); n != 1 {
		t.Errorf("chunk count = %d, want 1", n)
	}
}

func put(t *testing.T, p *LevelDB, key string, value []byte) {
	t.Helper()
	if err := p.db.Put([]byte(key), value, nil); err != nil {
		t.Fatal(err)
	}
}

// legacyIdMetaSubChunk builds 4096 IDs + 2048 meta nibbles in XZY order, filled with id:meta.
func legacyIdMeta(size int, id, meta byte) ([]byte, []byte) {
	ids := make([]byte, size)
	data := make([]byte, size/2)
	for i := range ids {
		ids[i] = id
	}
	for i := range data {
		data[i] = meta | meta<<4
	}
	return ids, data
}

// A 1.1-era chunk: classic (version 0) subchunks with legacy numeric IDs, 2D biomes.
func TestLoadsLegacyClassicSubChunks(t *testing.T) {
	p := newProvider(t)
	index := ChunkIndex(0, 0)
	put(t, p, index+ChunkDataKeyNewVersion, []byte{ChunkVersionV1_1_0})
	ids, meta := legacyIdMeta(4096, 5, 2) // birch planks
	put(t, p, index+ChunkDataKeySubChunk+"\x00", append(append([]byte{SubChunkVersionClassic}, ids...), meta...))
	biomes := make([]byte, 512+256)
	for i := 512; i < len(biomes); i++ {
		biomes[i] = 4 // forest
	}
	put(t, p, index+ChunkDataKeyHeightmapAnd2DBiomes, biomes)

	loaded, err := p.LoadChunk(0, 0)
	if err != nil || loaded == nil {
		t.Fatalf("LoadChunk: %v %v", loaded, err)
	}
	if !loaded.IsUpgraded() {
		t.Errorf("a legacy chunk should be marked as upgraded")
	}
	sc := loaded.GetData().GetSubChunks()[0]
	if got := stateName(t, sc.GetBlockStateID(7, 7, 7)); got != "Birch Planks" {
		t.Errorf("block = %s, want Birch Planks", got)
	}
	if got := loaded.GetData().GetSubChunks()[10].GetBiomeArray().Get(3, 3, 3); got != 4 {
		t.Errorf("extrapolated biome = %d, want 4", got)
	}
	if !loaded.GetData().IsPopulated() {
		t.Errorf("chunks without a finalization tag count as populated")
	}
}

// 3D biomes may use the 127 marker to copy the previous subchunk's palette.
func TestLoads3DBiomesWithCopyMarker(t *testing.T) {
	p := newProvider(t)
	index := ChunkIndex(1, 1)
	put(t, p, index+ChunkDataKeyNewVersion, []byte{ChunkVersionV1_18_30})
	stream := binaryutils.NewBinaryStream(make([]byte, 512), 512)
	serializeBiomePalette(stream, format.NewPalettedBlockArray(7))
	for i := 1; i < 24; i++ {
		stream.PutByte(127 << 1)
	}
	put(t, p, index+ChunkDataKeyHeightmapAnd3DBiomes, stream.GetBuffer())

	loaded, err := p.LoadChunk(1, 1)
	if err != nil || loaded == nil {
		t.Fatalf("LoadChunk: %v %v", loaded, err)
	}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		if got := loaded.GetData().GetSubChunks()[y].GetBiomeArray().Get(0, 0, 0); got != 7 {
			t.Fatalf("subchunk %d biome = %d, want 7", y, got)
		}
	}
}

// An MCPE 0.9 chunk: the whole 16x128x16 column in LEGACY_TERRAIN.
func TestLoadsLegacyTerrain(t *testing.T) {
	p := newProvider(t)
	index := ChunkIndex(2, 2)
	put(t, p, index+ChunkDataKeyOldVersion, []byte{ChunkVersionV0_9_5})
	ids, meta := legacyIdMeta(32768, 1, 1) // granite
	terrain := append(append(ids, meta...), make([]byte, 32768+256+1024)...)
	put(t, p, index+ChunkDataKeyLegacyTerrain, terrain)

	loaded, err := p.LoadChunk(2, 2)
	if err != nil || loaded == nil {
		t.Fatalf("LoadChunk: %v %v", loaded, err)
	}
	if got := stateName(t, loaded.GetData().GetSubChunks()[7].GetBlockStateID(0, 15, 0)); got != "Granite" {
		t.Errorf("block = %s, want Granite", got)
	}
	if got := loaded.GetData().GetSubChunks()[8].GetBlockStateID(0, 0, 0); got != io.EmptyStateID() {
		t.Errorf("above y=128 should be empty")
	}
}

func TestCorruptedChunks(t *testing.T) {
	p := newProvider(t)
	put(t, p, ChunkIndex(3, 3)+ChunkDataKeyNewVersion, []byte{200})
	_, err := p.LoadChunk(3, 3)
	var corrupted *exception.CorruptedChunkError
	if !errors.As(err, &corrupted) {
		t.Errorf("unknown chunk version: got %v, want CorruptedChunkError", err)
	}

	index := ChunkIndex(4, 4)
	put(t, p, index+ChunkDataKeyNewVersion, []byte{CurrentLevelChunkVersion})
	put(t, p, index+ChunkDataKeySubChunk+"\x00", []byte{SubChunkVersionPalettedMulti, 1, 3 << 1}) // truncated
	if _, err := p.LoadChunk(4, 4); !errors.As(err, &corrupted) {
		t.Errorf("truncated subchunk: got %v, want CorruptedChunkError", err)
	}

	// GetAllChunks skips corrupted chunks when asked to.
	count := 0
	if err := p.GetAllChunks(true, nil, func(io.ChunkCoords, *io.LoadedChunkData) bool { count++; return true }); err != nil {
		t.Fatal(err)
	}
	if err := p.GetAllChunks(false, nil, func(io.ChunkCoords, *io.LoadedChunkData) bool { return true }); err == nil {
		t.Errorf("GetAllChunks without skipCorrupted should fail")
	}
}
