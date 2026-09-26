package region

import (
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
)

// Anvil is a port of pocketmine\world\format\io\region\Anvil: the (pre-1.13, numeric block ID)
// Java Edition Anvil format.
type Anvil struct {
	*RegionWorldProvider
}

var _ io.WorldProvider = (*Anvil)(nil)

var anvilFormat = regionFormat{
	extension:       "mca",
	pcFormatVersion: 19133,
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
			// ignore legacy light information
			return format.NewSubChunk(io.EmptyStateID(), []*format.PalettedBlockArray{p.PalettizeLegacySubChunkYZX(blocks, meta, logger)}, biomes3d), nil
		})
	},
}

// NewAnvil is a port of `new Anvil($path, $logger)`.
func NewAnvil(path string, logger log.Logger) (*Anvil, error) {
	p, err := newRegionWorldProvider(path, logger, anvilFormat)
	if err != nil {
		return nil, err
	}
	return &Anvil{p}, nil
}

// IsValidAnvil is Anvil::isValid.
func IsValidAnvil(path string) bool { return isValidRegionWorld(path, anvilFormat.extension) }

// GetWorldMinY is a port of Anvil::getWorldMinY.
func (p *Anvil) GetWorldMinY() int { return 0 }

// GetWorldMaxY is a port of Anvil::getWorldMaxY.
func (p *Anvil) GetWorldMaxY() int {
	//TODO: add world height options
	return 256
}
