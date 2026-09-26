package region

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format/io"
	worlddata "pocketmine-go/pocketmine/world/format/io/data"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// regionFormat is the per-format part of a RegionWorldProvider (PHP's abstract static methods and
// deserializeChunk).
type regionFormat struct {
	extension       string // getRegionFileExtension
	pcFormatVersion int    // getPcWorldFormatVersion
	deserialize     func(p *RegionWorldProvider, data []byte, logger log.Logger) (*io.LoadedChunkData, error)
}

// RegionWorldProvider is a port of pocketmine\world\format\io\region\RegionWorldProvider.
type RegionWorldProvider struct {
	io.BaseWorldProvider
	format  regionFormat
	regions map[[2]int]*RegionLoader
}

// isValidRegionWorld is a port of RegionWorldProvider::isValid: a level.dat and a region directory
// with at least one file of this format's extension.
func isValidRegionWorld(path, extension string) bool {
	if _, err := os.Stat(filepath.Join(path, "level.dat")); err != nil {
		return false
	}
	regionPath := filepath.Join(path, "region")
	info, err := os.Stat(regionPath)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(regionPath)
	if err != nil {
		// we can't tell the type if we don't have read perms
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if extPos := strings.LastIndex(name, "."); extPos != -1 && name[extPos+1:] == extension {
			// we don't care if other region types exist, we only care if this format is possible
			return true
		}
	}
	return false
}

func newRegionWorldProvider(path string, logger log.Logger, format regionFormat) (*RegionWorldProvider, error) {
	p := &RegionWorldProvider{format: format, regions: map[[2]int]*RegionLoader{}}
	if err := p.InitBase(path, logger, func() (io.WorldData, error) {
		return worlddata.LoadJavaWorldData(filepath.Join(path, "level.dat"))
	}); err != nil {
		return nil, err
	}
	return p, nil
}

// DoGarbageCollection is a port of RegionWorldProvider::doGarbageCollection: closes regions unused
// for 5 minutes.
func (p *RegionWorldProvider) DoGarbageCollection() {
	limit := time.Now().Unix() - 300
	for index, region := range p.regions {
		if region.LastUsed <= limit {
			_ = region.Close()
			delete(p.regions, index)
		}
	}
}

// GetRegionIndex is a port of RegionWorldProvider::getRegionIndex.
func GetRegionIndex(chunkX, chunkZ int) (regionX, regionZ int) {
	return chunkX >> 5, chunkZ >> 5
}

// pathToRegion is a port of RegionWorldProvider::pathToRegion.
func (p *RegionWorldProvider) pathToRegion(regionX, regionZ int) string {
	return filepath.Join(p.Path, "region", fmt.Sprintf("r.%d.%d.%s", regionX, regionZ, p.format.extension))
}

// loadRegion is a port of RegionWorldProvider::loadRegion: a corrupted region file is backed up
// and replaced with an empty one.
func (p *RegionWorldProvider) loadRegion(regionX, regionZ int) (*RegionLoader, error) {
	index := [2]int{regionX, regionZ}
	if region, ok := p.regions[index]; ok {
		return region, nil
	}
	path := p.pathToRegion(regionX, regionZ)
	region, err := LoadExisting(path)
	if err != nil {
		var corrupted *CorruptedRegionError
		if !errors.As(err, &corrupted) {
			return nil, err
		}
		p.Logger.Error("Corrupted region file detected: " + corrupted.Message)
		backupPath := fmt.Sprintf("%s.bak.%d", path, time.Now().Unix())
		_ = os.Rename(path, backupPath)
		p.Logger.Error("Corrupted region file has been backed up to " + backupPath)
		if region, err = CreateNew(path); err != nil {
			return nil, err
		}
	}
	p.regions[index] = region
	return region, nil
}

// unloadRegion is a port of RegionWorldProvider::unloadRegion.
func (p *RegionWorldProvider) unloadRegion(regionX, regionZ int) {
	index := [2]int{regionX, regionZ}
	if region, ok := p.regions[index]; ok {
		_ = region.Close()
		delete(p.regions, index)
	}
}

// Close is a port of RegionWorldProvider::close.
func (p *RegionWorldProvider) Close() error {
	for index, region := range p.regions {
		_ = region.Close()
		delete(p.regions, index)
	}
	return nil
}

// getCompoundList is a port of RegionWorldProvider::getCompoundList.
func getCompoundList(context string, list *nbt.ListTag) ([]*nbt.CompoundTag, error) {
	if list.Count() > 0 && list.GetTagType() != nbt.TagCompound {
		return nil, exception.NewCorruptedChunkError("Expected TAG_List<TAG_Compound> for '%s'", context)
	}
	result := make([]*nbt.CompoundTag, 0, list.Count())
	for _, t := range list.Values() {
		result = append(result, t.(*nbt.CompoundTag))
	}
	return result, nil
}

// readFixedSizeByteArray is a port of RegionWorldProvider::readFixedSizeByteArray.
func readFixedSizeByteArray(chunk *nbt.CompoundTag, tagName string, length int) ([]byte, error) {
	t, ok := chunk.GetTag(tagName)
	if !ok {
		return nil, exception.NewCorruptedChunkError("'%s' key is missing from chunk NBT", tagName)
	}
	arr, ok := t.(nbt.ByteArrayTag)
	if !ok {
		return nil, exception.NewCorruptedChunkError("Expected TAG_ByteArray for '%s'", tagName)
	}
	if len(arr) != length {
		return nil, exception.NewCorruptedChunkError("Expected '%s' payload to have exactly %d bytes, but have %d", tagName, length, len(arr))
	}
	return []byte(arr), nil
}

// LoadChunk is a port of RegionWorldProvider::loadChunk.
func (p *RegionWorldProvider) LoadChunk(chunkX, chunkZ int) (*io.LoadedChunkData, error) {
	regionX, regionZ := GetRegionIndex(chunkX, chunkZ)
	if _, err := os.Stat(p.pathToRegion(regionX, regionZ)); err != nil {
		return nil, nil
	}
	region, err := p.loadRegion(regionX, regionZ)
	if err != nil {
		return nil, err
	}
	chunkData, err := region.ReadChunk(chunkX&0x1f, chunkZ&0x1f)
	if err != nil {
		return nil, err
	}
	if chunkData == nil {
		return nil, nil
	}
	return p.format.deserialize(p, chunkData, log.NewPrefixedLogger(p.Logger, fmt.Sprintf("Loading chunk x=%d z=%d", chunkX, chunkZ)))
}

// regionFiles is RegionWorldProvider::createRegionIterator: the coordinates of every region file.
func (p *RegionWorldProvider) regionFiles() ([][2]int, error) {
	entries, err := os.ReadDir(filepath.Join(p.Path, "region"))
	if err != nil {
		return nil, err
	}
	pattern := regexp.MustCompile(`^r\.(-?\d+)\.(-?\d+)\.` + regexp.QuoteMeta(p.format.extension) + `$`)
	var result [][2]int
	for _, e := range entries {
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		x, _ := strconv.Atoi(m[1])
		z, _ := strconv.Atoi(m[2])
		result = append(result, [2]int{x, z})
	}
	return result, nil
}

// GetAllChunks is a port of RegionWorldProvider::getAllChunks.
func (p *RegionWorldProvider) GetAllChunks(skipCorrupted bool, logger log.Logger, yield func(coords io.ChunkCoords, chunk *io.LoadedChunkData) bool) error {
	regions, err := p.regionFiles()
	if err != nil {
		return err
	}
	for _, region := range regions {
		regionX, regionZ := region[0], region[1]
		rX, rZ := regionX<<5, regionZ<<5
		for chunkX := rX; chunkX < rX+32; chunkX++ {
			for chunkZ := rZ; chunkZ < rZ+32; chunkZ++ {
				chunk, err := p.LoadChunk(chunkX, chunkZ)
				if err != nil {
					var corrupted *exception.CorruptedChunkError
					if !errors.As(err, &corrupted) || !skipCorrupted {
						return err
					}
					if logger != nil {
						logger.Error(fmt.Sprintf("Skipped corrupted chunk %d %d (%v)", chunkX, chunkZ, err))
					}
					continue
				}
				if chunk != nil && !yield(io.ChunkCoords{X: chunkX, Z: chunkZ}, chunk) {
					return nil
				}
			}
		}
		p.unloadRegion(regionX, regionZ)
	}
	return nil
}

// CalculateChunkCount is a port of RegionWorldProvider::calculateChunkCount.
func (p *RegionWorldProvider) CalculateChunkCount() (int, error) {
	regions, err := p.regionFiles()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, region := range regions {
		loader, err := p.loadRegion(region[0], region[1])
		if err != nil {
			return 0, err
		}
		count += loader.CalculateChunkCount()
		p.unloadRegion(region[0], region[1])
	}
	return count, nil
}
