package region

import (
	"fmt"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
	worlddata "pocketmine-go/pocketmine/world/format/io/data"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// biomeOcean is BiomeIds::OCEAN.
const biomeOcean = 0

// readLevelTag decompresses a region chunk and returns its "Level" compound.
func readLevelTag(data []byte) (*nbt.CompoundTag, error) {
	decompressed, err := worlddata.ZlibDecode(data)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to decompress chunk NBT")
	}
	root, _, err := nbt.NewBigEndianSerializer().Read(decompressed, 0, 512)
	if err != nil {
		return nil, exception.WrapCorruptedChunk(err)
	}
	tag, err := root.MustGetCompoundTag()
	if err != nil {
		return nil, exception.WrapCorruptedChunk(err)
	}
	levelTag, _ := tag.GetTag("Level")
	level, ok := levelTag.(*nbt.CompoundTag)
	if !ok {
		return nil, exception.NewCorruptedChunkError("'Level' key is missing from chunk NBT")
	}
	return level, nil
}

// makeBiomeArray is the $makeBiomeArray closure of LegacyAnvilChunkTrait/McRegion.
func makeBiomeArray(biomeIDs []byte) (*format.PalettedBlockArray, error) {
	if len(biomeIDs) != 256 {
		return nil, exception.NewCorruptedChunkError("Expected biome array to be exactly 256 bytes, got %d", len(biomeIDs))
	}
	//TODO: we may need to convert legacy biome IDs
	return io.Extrapolate3DBiomes(biomeIDs), nil
}

// readBiomes reads BiomeColors (MCPE 0.9) or Biomes, falling back to ocean.
func readBiomes(chunk *nbt.CompoundTag) (*format.PalettedBlockArray, error) {
	if t, ok := chunk.GetTag("BiomeColors"); ok {
		if colors, ok := t.(nbt.IntArrayTag); ok {
			return makeBiomeArray(io.ConvertBiomeColors([]int32(colors))) // Convert back to original format
		}
	}
	if t, ok := chunk.GetTag("Biomes"); ok {
		if biomes, ok := t.(nbt.ByteArrayTag); ok {
			return makeBiomeArray([]byte(biomes))
		}
	}
	return format.NewPalettedBlockArray(biomeOcean), nil
}

// chunkEntitiesAndTiles reads the "Entities" and "TileEntities" lists.
func chunkEntitiesAndTiles(chunk *nbt.CompoundTag) ([]*nbt.CompoundTag, []*nbt.CompoundTag, error) {
	var entities, tiles []*nbt.CompoundTag
	var err error
	if t, ok := chunk.GetTag("Entities"); ok {
		if list, ok := t.(*nbt.ListTag); ok {
			if entities, err = getCompoundList("Entities", list); err != nil {
				return nil, nil, err
			}
		}
	}
	if t, ok := chunk.GetTag("TileEntities"); ok {
		if list, ok := t.(*nbt.ListTag); ok {
			if tiles, err = getCompoundList("TileEntities", list); err != nil {
				return nil, nil, err
			}
		}
	}
	return entities, tiles, nil
}

// deserializeLegacyAnvilChunk is a port of LegacyAnvilChunkTrait::deserializeChunk.
// deserializeSubChunk is the format's deserializeSubChunk.
func deserializeLegacyAnvilChunk(data []byte, logger log.Logger, deserializeSubChunk func(subChunk *nbt.CompoundTag, biomes3d *format.PalettedBlockArray, logger log.Logger) (*format.SubChunk, error)) (*io.LoadedChunkData, error) {
	chunk, err := readLevelTag(data)
	if err != nil {
		return nil, err
	}
	biomes3d, err := readBiomes(chunk)
	if err != nil {
		return nil, err
	}

	subChunks := map[int]*format.SubChunk{}
	if sections, ok, _ := chunk.GetListTag("Sections"); ok {
		for _, t := range sections.Values() {
			subChunk, ok := t.(*nbt.CompoundTag)
			if !ok {
				return nil, exception.NewCorruptedChunkError("Expected TAG_List<TAG_Compound> for 'Sections'")
			}
			y := int(subChunk.GetByteOr("Y", 0))
			sc, err := deserializeSubChunk(subChunk, biomes3d.Clone(), log.NewPrefixedLogger(logger, fmt.Sprintf("Subchunk y=%d", y)))
			if err != nil {
				return nil, err
			}
			subChunks[y] = sc
		}
	}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		if _, ok := subChunks[y]; !ok {
			subChunks[y] = format.NewSubChunk(io.EmptyStateID(), nil, biomes3d.Clone())
		}
	}

	entities, tiles, err := chunkEntitiesAndTiles(chunk)
	if err != nil {
		return nil, err
	}
	return io.NewLoadedChunkData(
		io.NewChunkData(subChunks, chunk.GetByteOr("TerrainPopulated", 0) != 0, entities, tiles),
		true,
		io.FixerFlagAll,
	), nil
}
