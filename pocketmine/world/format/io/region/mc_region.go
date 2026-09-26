package region

import (
	"fmt"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
)

// McRegion is a port of pocketmine\world\format\io\region\McRegion: the Java Edition Beta /
// early PocketMine-MP format (128-block-high columns).
type McRegion struct {
	*RegionWorldProvider
}

var _ io.WorldProvider = (*McRegion)(nil)

var mcRegionFormat = regionFormat{
	extension:       "mcr",
	pcFormatVersion: 19132,
	deserialize:     deserializeMcRegionChunk,
}

// deserializeMcRegionChunk is a port of McRegion::deserializeChunk.
func deserializeMcRegionChunk(p *RegionWorldProvider, data []byte, logger log.Logger) (*io.LoadedChunkData, error) {
	chunk, err := readLevelTag(data)
	if err != nil {
		return nil, err
	}
	if t, ok := chunk.GetTag("TerrainGenerated"); ok {
		if generated, ok := t.(nbt.ByteTag); ok && generated == 0 {
			// In legacy PM before 3.0, PM used to save MCRegion chunks even when they weren't
			// generated. In these cases (we'll see them in old worlds), some of the tags which we
			// expect to always be present, will be missing. If TerrainGenerated (PM-specific tag from
			// the olden days) is false, toss the chunk data and don't bother trying to read it.
			return nil, nil
		}
	}
	biomes3d, err := readBiomes(chunk)
	if err != nil {
		return nil, err
	}

	subChunks := map[int]*format.SubChunk{}
	fullIds, err := readFixedSizeByteArray(chunk, "Blocks", 32768)
	if err != nil {
		return nil, err
	}
	fullData, err := readFixedSizeByteArray(chunk, "Data", 16384)
	if err != nil {
		return nil, err
	}
	for y := 0; y < 8; y++ {
		layer := p.PalettizeLegacySubChunkFromColumn(fullIds, fullData, y, log.NewPrefixedLogger(logger, fmt.Sprintf("Subchunk y=%d", y)))
		subChunks[y] = format.NewSubChunk(io.EmptyStateID(), []*format.PalettedBlockArray{layer}, biomes3d.Clone())
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

// NewMcRegion is a port of `new McRegion($path, $logger)`.
func NewMcRegion(path string, logger log.Logger) (*McRegion, error) {
	p, err := newRegionWorldProvider(path, logger, mcRegionFormat)
	if err != nil {
		return nil, err
	}
	return &McRegion{p}, nil
}

// IsValidMcRegion is McRegion::isValid.
func IsValidMcRegion(path string) bool { return isValidRegionWorld(path, mcRegionFormat.extension) }

// GetWorldMinY is a port of McRegion::getWorldMinY.
func (p *McRegion) GetWorldMinY() int { return 0 }

// GetWorldMaxY is a port of McRegion::getWorldMaxY.
func (p *McRegion) GetWorldMaxY() int {
	//TODO: add world height options
	return 128
}
