package leveldb

import (
	"fmt"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/world/format"
)

// serializeBiomePalette is a port of LevelDB::serializeBiomePalette - the same envelope shape as a
// block layer, but palette values are raw (already-numeric) biome IDs written directly as
// little-endian ints, never NBT (biomes have no persistent name/states form in this scheme).
func serializeBiomePalette(biomes *format.PalettedBlockArray) []byte {
	bitsPerBlock := biomes.GetBitsPerBlock()
	buf := []byte{byte(bitsPerBlock << 1)}
	buf = append(buf, biomes.GetWordArray()...)

	palette := biomes.GetPalette()
	if bitsPerBlock != 0 {
		buf = append(buf, binaryutils.WriteLInt(int32(len(palette)))...)
	}
	for _, biomeID := range palette {
		buf = append(buf, binaryutils.WriteLInt(biomeID)...)
	}
	return buf
}

// deserializeBiomePalette is the reverse of serializeBiomePalette, returning the new offset past
// what it consumed (this port doesn't support the vanilla "reuse the previous subchunk's palette"
// shorthand real deserialize3dBiomes handles via a bitsPerBlock=127 marker - see this package's
// doc comment: only round-tripping worlds this port itself saved is in scope, and this port's own
// writer never emits that shorthand).
func deserializeBiomePalette(data []byte, offset int) (*format.PalettedBlockArray, int, error) {
	if offset >= len(data) {
		return nil, 0, fmt.Errorf("leveldb: biome data truncated before palette header")
	}
	bitsPerBlock := int(data[offset]) >> 1
	offset++

	var wordBytes []byte
	if bitsPerBlock != 0 {
		wordLen := format.WordCountForBitsPerBlock(bitsPerBlock) * 4
		if offset+wordLen > len(data) {
			return nil, 0, fmt.Errorf("leveldb: biome data truncated in word array")
		}
		wordBytes = data[offset : offset+wordLen]
		offset += wordLen
	}

	paletteCount := 1
	if bitsPerBlock != 0 {
		if offset+4 > len(data) {
			return nil, 0, fmt.Errorf("leveldb: biome data truncated before palette count")
		}
		n, err := binaryutils.ReadLInt(data[offset : offset+4])
		if err != nil {
			return nil, 0, err
		}
		paletteCount = int(n)
		offset += 4
	}

	if offset+paletteCount*4 > len(data) {
		return nil, 0, fmt.Errorf("leveldb: biome data truncated in palette values")
	}
	palette := make([]int32, paletteCount)
	for i := 0; i < paletteCount; i++ {
		v, err := binaryutils.ReadLInt(data[offset : offset+4])
		if err != nil {
			return nil, 0, err
		}
		palette[i] = v
		offset += 4
	}

	arr, err := format.NewPalettedBlockArrayFromRaw(bitsPerBlock, wordBytes, palette)
	if err != nil {
		return nil, 0, err
	}
	return arr, offset, nil
}
