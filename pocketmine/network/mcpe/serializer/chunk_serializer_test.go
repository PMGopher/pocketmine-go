package serializer

import (
	"testing"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/format"
)

func newTestAir() *block.Air {
	idInfo, err := block.NewBlockIdentifier(block.AIR, nil)
	if err != nil {
		panic(err)
	}
	return block.NewAir(idInfo, "Air", block.NewBlockTypeInfo(block.BlockBreakInfoIndestructible(-1.0), nil, nil))
}

func newTestStone() *block.Stone {
	idInfo, err := block.NewBlockIdentifier(block.STONE, nil)
	if err != nil {
		panic(err)
	}
	return block.NewStone(idInfo, "Stone", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))
}

func TestSerializeBiomePaletteUniform(t *testing.T) {
	biomes := format.NewPalettedBlockArray(1)

	got := serializeBiomePalette(biomes)

	// header byte: (bitsPerBlock=0 << 1) | 1 = 1; no words; no count (bitsPerBlock is 0); one
	// palette value, zigzag(1) = 2, fits in a single VarInt byte.
	want := append([]byte{1}, binaryutils.WriteVarInt(1)...)
	if string(got) != string(want) {
		t.Errorf("serializeBiomePalette(uniform) = %v, want %v", got, want)
	}
}

func TestSerializeBiomePaletteWithMultipleValues(t *testing.T) {
	biomes := format.NewPalettedBlockArray(1)
	biomes.Set(0, 0, 0, 4)

	got := serializeBiomePalette(biomes)

	offset := 0
	header := got[offset]
	offset++
	bitsPerBlock := int(header >> 1)
	if header&1 != 1 {
		t.Error("expected the non-persistence bit to be set")
	}
	if bitsPerBlock != 1 {
		t.Fatalf("bitsPerBlock = %d, want 1", bitsPerBlock)
	}
	blocksPerWord := 32 / bitsPerBlock
	wordCount := (4096 + blocksPerWord - 1) / blocksPerWord
	offset += wordCount * 4

	count, err := binaryutils.ReadVarInt(got, &offset)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("palette count = %d, want 2", count)
	}
	first, err := binaryutils.ReadVarInt(got, &offset)
	if err != nil {
		t.Fatal(err)
	}
	second, err := binaryutils.ReadVarInt(got, &offset)
	if err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 4 {
		t.Errorf("palette = [%d, %d], want [1, 4]", first, second)
	}
	if offset != len(got) {
		t.Errorf("expected to consume exactly the whole buffer, %d bytes left over", len(got)-offset)
	}
}

func TestGetSubChunkCountIsZeroForAllEmptyChunk(t *testing.T) {
	c := format.NewChunk(nil, false, 0, 1)
	if got := GetSubChunkCount(c); got != 0 {
		t.Errorf("GetSubChunkCount() = %d, want 0", got)
	}
}

func TestGetSubChunkCountCountsUpToTopmostNonEmpty(t *testing.T) {
	c := format.NewChunk(nil, false, 0, 1)
	// Populate the subchunk at Y index 2.
	c.GetSubChunk(2).SetBlockStateID(0, 0, 0, 5)

	want := 2 - format.MinSubChunkIndex + 1
	if got := GetSubChunkCount(c); got != want {
		t.Errorf("GetSubChunkCount() = %d, want %d", got, want)
	}
}

func TestSerializeSubChunkStructure(t *testing.T) {
	tr := convert.NewBlockTranslator()
	air := newTestAir()
	stone := newTestStone()
	airNetID := tr.InternalIDToNetworkID(air)
	stoneNetID := tr.InternalIDToNetworkID(stone)

	sc := format.NewSubChunk(int32(air.GetStateId()), nil, format.NewPalettedBlockArray(1))
	sc.SetBlockStateID(0, 0, 0, int32(stone.GetStateId()))

	data := SerializeSubChunk(sc, tr)

	offset := 0
	if data[offset] != 8 {
		t.Fatalf("version byte = %d, want 8", data[offset])
	}
	offset++
	layerCount := data[offset]
	offset++
	if layerCount != 1 {
		t.Fatalf("layer count = %d, want 1", layerCount)
	}

	header := data[offset]
	offset++
	bitsPerBlock := int(header >> 1)
	if header&1 != 1 {
		t.Errorf("expected the non-persistence bit to be set")
	}
	if bitsPerBlock != 1 { // 2 distinct values (air, stone) -> 1 bit
		t.Fatalf("bitsPerBlock = %d, want 1", bitsPerBlock)
	}

	blocksPerWord := 32 / bitsPerBlock
	wordCount := (4096 + blocksPerWord - 1) / blocksPerWord
	offset += wordCount * 4 // skip the raw word array - PalettedBlockArray's own tests already cover its correctness

	paletteCount, err := binaryutils.ReadVarInt(data, &offset)
	if err != nil {
		t.Fatalf("failed to read palette count: %v", err)
	}
	if paletteCount != 2 {
		t.Fatalf("palette count = %d, want 2", paletteCount)
	}

	first, err := binaryutils.ReadVarInt(data, &offset)
	if err != nil {
		t.Fatal(err)
	}
	second, err := binaryutils.ReadVarInt(data, &offset)
	if err != nil {
		t.Fatal(err)
	}

	if first != airNetID {
		t.Errorf("first palette entry (network ID) = %d, want air's %d", first, airNetID)
	}
	if second != stoneNetID {
		t.Errorf("second palette entry (network ID) = %d, want stone's %d", second, stoneNetID)
	}
	if offset != len(data) {
		t.Errorf("expected to consume exactly the whole buffer, %d bytes left over", len(data)-offset)
	}
}

func TestSerializeFullChunkEndsWithBorderBlocksAndEmptyTiles(t *testing.T) {
	tr := convert.NewBlockTranslator()
	c := format.NewChunk(nil, false, int32(newTestAir().GetStateId()), 1)

	data := SerializeFullChunk(c, tr)

	// No populated subchunks -> 0 subchunk sections, then one biome palette per subchunk slot,
	// then a 0 border block count byte, then a 0-length (VarUint) tiles section.
	var want []byte
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		want = append(want, serializeBiomePalette(c.GetSubChunk(y).GetBiomeArray())...)
	}
	want = append(want, 0)
	want = append(want, binaryutils.WriteUnsignedVarInt(0)...)

	if string(data) != string(want) {
		t.Errorf("len(data) = %d, want %d (mismatched bytes)", len(data), len(want))
	}
}

func TestSerializeFullChunkIncludesEverySubChunkUpToTopmostNonEmpty(t *testing.T) {
	tr := convert.NewBlockTranslator()
	air := newTestAir()
	c := format.NewChunk(nil, false, int32(air.GetStateId()), 1)
	c.GetSubChunk(0).SetBlockStateID(0, 0, 0, int32(newTestStone().GetStateId()))

	data := SerializeFullChunk(c, tr)

	expectedSubChunkCount := 0 - format.MinSubChunkIndex + 1
	if got := GetSubChunkCount(c); got != expectedSubChunkCount {
		t.Fatalf("GetSubChunkCount() = %d, want %d", got, expectedSubChunkCount)
	}

	var want []byte
	writtenCount := 0
	for y := format.MinSubChunkIndex; writtenCount < expectedSubChunkCount; y, writtenCount = y+1, writtenCount+1 {
		want = append(want, SerializeSubChunk(c.GetSubChunk(y), tr)...)
	}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		want = append(want, serializeBiomePalette(c.GetSubChunk(y).GetBiomeArray())...)
	}
	want = append(want, 0)
	want = append(want, binaryutils.WriteUnsignedVarInt(0)...)

	if string(data) != string(want) {
		t.Errorf("len(data) = %d, want %d (mismatched bytes)", len(data), len(want))
	}
}
