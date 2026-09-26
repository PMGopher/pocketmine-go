package region

import (
	"os"
	"path/filepath"

	"pocketmine-go/pocketmine/world/format/io"
	worlddata "pocketmine-go/pocketmine/world/format/io/data"
)

// WritableRegionWorldProvider is a port of pocketmine\world\format\io\region\
// WritableRegionWorldProvider: a region provider that can save chunks. PocketMine-MP has no
// concrete writable region format any more (region worlds are converted to LevelDB); a format
// provides serializeChunk to use it.
type WritableRegionWorldProvider struct {
	*RegionWorldProvider
	// serializeChunk is the format's serializeChunk: zlib-compressed chunk NBT.
	serializeChunk func(chunk *io.ChunkData) ([]byte, error)
}

// GenerateRegionWorld is a port of WritableRegionWorldProvider::generate.
func GenerateRegionWorld(path, name string, options io.WorldCreationOptions, pcFormatVersion int) error {
	if err := os.MkdirAll(filepath.Join(path, "region"), 0o777); err != nil {
		return err
	}
	return worlddata.GenerateJavaWorldData(path, name, options, pcFormatVersion)
}

// SaveChunk is a port of WritableRegionWorldProvider::saveChunk.
func (p *WritableRegionWorldProvider) SaveChunk(chunkX, chunkZ int, chunkData *io.ChunkData, dirtyFlags int) error {
	regionX, regionZ := GetRegionIndex(chunkX, chunkZ)
	region, err := p.loadRegion(regionX, regionZ)
	if err != nil {
		return err
	}
	data, err := p.serializeChunk(chunkData)
	if err != nil {
		return err
	}
	return region.WriteChunk(chunkX&0x1f, chunkZ&0x1f, data)
}
