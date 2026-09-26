package io

import "pocketmine-go/pocketmine/world/format"

// SubChunkConverter is a port of the SubChunkConverter class from PocketMine-MP's ext-chunkutils2
// C++ extension: it palettizes legacy (pre-palette) block ID + meta arrays. Each value is
// (id << 4) | meta; BaseWorldProvider then upgrades the palette to current block states.
//
// Meta nibbles are packed two per byte, the low nibble first.

func legacyNibble(meta []byte, index int) int32 {
	return int32(meta[index>>1]>>((index&1)<<2)) & 0xf
}

func convertLegacy(ids, meta []byte, index func(x, y, z int) int) *format.PalettedBlockArray {
	result := format.NewPalettedBlockArray(0)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 0; y < 16; y++ {
				i := index(x, y, z)
				result.Set(x, y, z, int32(ids[i])<<4|legacyNibble(meta, i))
			}
		}
	}
	return result
}

// ConvertSubChunkXZY converts a 16x16x16 subchunk stored in XZY order (MCPE 1.0-1.2, PMAnvil).
func ConvertSubChunkXZY(idArray, metaArray []byte) *format.PalettedBlockArray {
	return convertLegacy(idArray, metaArray, func(x, y, z int) int { return x<<8 | z<<4 | y })
}

// ConvertSubChunkYZX converts a 16x16x16 subchunk stored in YZX order (Anvil).
func ConvertSubChunkYZX(idArray, metaArray []byte) *format.PalettedBlockArray {
	return convertLegacy(idArray, metaArray, func(x, y, z int) int { return y<<8 | z<<4 | x })
}

// ConvertSubChunkFromLegacyColumn converts subchunk yOffset of a 16x128x16 column stored in XZY
// order (MCPE 0.9 LevelDB, McRegion).
func ConvertSubChunkFromLegacyColumn(idArray, metaArray []byte, yOffset int) *format.PalettedBlockArray {
	return convertLegacy(idArray, metaArray, func(x, y, z int) int { return x<<11 | z<<7 | (yOffset<<4 + y) })
}
