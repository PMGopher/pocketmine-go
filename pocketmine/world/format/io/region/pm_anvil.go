package region

import (
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
)

// PMAnvil is a port of pocketmine\world\format\io\region\PMAnvil: Anvil with XZY-ordered sections,
// used by old PocketMine-MP versions.
type PMAnvil struct {
	*RegionWorldProvider
}

var _ io.WorldProvider = (*PMAnvil)(nil)

var pmAnvilFormat = regionFormat{
	extension:       "mcapm",
	pcFormatVersion: -1, // Not a PC format, only PocketMine-MP
	deserialize: func(p *RegionWorldProvider, data []byte, logger log.Logger) (*io.LoadedChunkData, error) {
		return deserializeLegacyAnvilChunk(data, logger, func(subChunk *nbt.CompoundTag, biomes3d *format.PalettedBlockArray, logger log.Logger) (*format.SubChunk, error) {
			blocks, err := readFixedSizeByteArray(subChunk, "Blocks", 4096)
			if err != nil {
				return nil, err
			}
			meta, err := readFixedSizeByteArray(subChunk, "Data", 2048)
			if err != nil {
				return nil, err
			}
			return format.NewSubChunk(io.EmptyStateID(), []*format.PalettedBlockArray{p.PalettizeLegacySubChunkXZY(blocks, meta, logger)}, biomes3d), nil
		})
	},
}

// NewPMAnvil is a port of `new PMAnvil($path, $logger)`.
func NewPMAnvil(path string, logger log.Logger) (*PMAnvil, error) {
	p, err := newRegionWorldProvider(path, logger, pmAnvilFormat)
	if err != nil {
		return nil, err
	}
	return &PMAnvil{p}, nil
}

// IsValidPMAnvil is PMAnvil::isValid.
func IsValidPMAnvil(path string) bool { return isValidRegionWorld(path, pmAnvilFormat.extension) }

// GetWorldMinY is a port of PMAnvil::getWorldMinY.
func (p *PMAnvil) GetWorldMinY() int { return 0 }

// GetWorldMaxY is a port of PMAnvil::getWorldMaxY.
func (p *PMAnvil) GetWorldMaxY() int { return 256 }
