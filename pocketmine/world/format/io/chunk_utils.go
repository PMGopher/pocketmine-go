package io

import (
	"fmt"

	"pocketmine-go/pocketmine/world/format"
)

// ConvertBiomeColors is a port of ChunkUtils::convertBiomeColors: MCPE 0.9-era biome colours (the
// biome ID in the top byte of each int) to a 256-byte biome ID array.
func ConvertBiomeColors(array []int32) []byte {
	result := make([]byte, 256)
	for i, color := range array {
		if i >= 256 {
			break
		}
		result[i] = byte(uint32(color) >> 24)
	}
	return result
}

// Extrapolate3DBiomes is a port of ChunkUtils::extrapolate3DBiomes: a 2D (ZX-ordered, 256 bytes)
// biome array repeated over the 16 blocks of a subchunk.
func Extrapolate3DBiomes(biomes2d []byte) *format.PalettedBlockArray {
	if len(biomes2d) != 256 {
		panic(fmt.Sprintf("Biome array is expected to be exactly 256 bytes, got %d", len(biomes2d)))
	}
	biomePalette := format.NewPalettedBlockArray(int32(biomes2d[0]))
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			biomeID := int32(biomes2d[z<<4|x])
			for y := 0; y < 16; y++ {
				biomePalette.Set(x, y, z, biomeID)
			}
		}
	}
	return biomePalette
}
