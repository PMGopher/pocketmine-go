package region

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"pocketmine-go/pocketmine/world/format/io/exception"
)

// RegionLoader compression types, a port of RegionLoader::COMPRESSION_*.
const (
	RegionCompressionGzip = 1
	RegionCompressionZlib = 2
)

const (
	regionMaxSectorLength = 255 << 12 // 255 sectors (~0.996 MiB)
	regionHeaderLength    = 8192      // 4096 location table + 4096 timestamps

	// RegionFirstSector is RegionLoader::FIRST_SECTOR: the location table occupies 0 and 1.
	RegionFirstSector = 2
)

// CorruptedRegionError is a port of CorruptedRegionException.
type CorruptedRegionError struct{ Message string }

func (e *CorruptedRegionError) Error() string { return e.Message }

// RegionLoader is a port of pocketmine\world\format\io\region\RegionLoader: reads and writes the
// chunks of one region file (32x32 chunks).
type RegionLoader struct {
	filePath      string
	file          *os.File
	nextSector    int
	locationTable []*RegionLocationTableEntry
	garbageTable  *RegionGarbageMap
	LastUsed      int64
}

func newRegionLoader(filePath string) (*RegionLoader, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &RegionLoader{
		filePath:     filePath,
		file:         f,
		nextSector:   RegionFirstSector,
		garbageTable: NewRegionGarbageMap(nil),
		LastUsed:     time.Now().Unix(),
	}, nil
}

// LoadExisting is a port of RegionLoader::loadExisting.
func LoadExisting(filePath string) (*RegionLoader, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("File %s does not exist", filePath)
	}
	if info.Size()%4096 != 0 {
		return nil, &CorruptedRegionError{"Region file should be padded to a multiple of 4KiB"}
	}
	r, err := newRegionLoader(filePath)
	if err != nil {
		return nil, err
	}
	if err := r.loadLocationTable(); err != nil {
		_ = r.Close()
		return nil, err
	}
	return r, nil
}

// CreateNew is a port of RegionLoader::createNew.
func CreateNew(filePath string) (*RegionLoader, error) {
	if _, err := os.Stat(filePath); err == nil {
		return nil, fmt.Errorf("Region file %s already exists", filePath)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	r, err := newRegionLoader(filePath)
	if err != nil {
		return nil, err
	}
	if err := r.createBlank(); err != nil {
		_ = r.Close()
		return nil, err
	}
	return r, nil
}

func (r *RegionLoader) isChunkGenerated(index int) bool { return r.locationTable[index] != nil }

// ReadChunk is a port of RegionLoader::readChunk: the chunk's compressed data, or nil if it isn't
// in the region.
func (r *RegionLoader) ReadChunk(x, z int) ([]byte, error) {
	index, err := getChunkOffset(x, z)
	if err != nil {
		return nil, err
	}
	r.LastUsed = time.Now().Unix()

	entry := r.locationTable[index]
	if entry == nil {
		return nil, nil
	}
	// this might cause us to read some junk, but under normal circumstances it won't be any more
	// than 4096 bytes wasted. doing this in a single call is faster than making two seeks and
	// reads to fetch the chunk. this relies on the assumption that the end of the file is always
	// padded to a multiple of 4096 bytes.
	bytesToRead := entry.GetSectorCount() << 12
	payload := make([]byte, bytesToRead)
	n, err := r.file.ReadAt(payload, int64(entry.GetFirstSector())<<12)
	if n != bytesToRead {
		return nil, exception.NewCorruptedChunkError("Corrupted chunk detected (unexpected EOF, truncated or non-padded chunk found)")
	}
	_ = err

	length := int32(binary.BigEndian.Uint32(payload[0:4]))
	if length <= 0 { //TODO: if we reached here, the locationTable probably needs updating
		return nil, nil
	}
	compression := payload[4]
	if compression != RegionCompressionZlib && compression != RegionCompressionGzip {
		return nil, exception.NewCorruptedChunkError("Invalid compression type (got %d, expected %d or %d)", compression, RegionCompressionZlib, RegionCompressionGzip)
	}
	end := 5 + int(length) - 1 // length prefix includes the compression byte
	if end > len(payload) {
		return nil, exception.NewCorruptedChunkError("Corrupted chunk detected: not enough bytes left")
	}
	return payload[5:end], nil
}

// ChunkExists is a port of RegionLoader::chunkExists.
func (r *RegionLoader) ChunkExists(x, z int) (bool, error) {
	index, err := getChunkOffset(x, z)
	if err != nil {
		return false, err
	}
	return r.isChunkGenerated(index), nil
}

func (r *RegionLoader) disposeGarbageArea(oldLocation *RegionLocationTableEntry) error {
	// release the area containing the old copy to the garbage pool
	if err := r.garbageTable.Add(oldLocation); err != nil {
		return err
	}
	nextSector := r.nextSector
	for endGarbage := r.garbageTable.End(); endGarbage != nil && endGarbage.GetLastSector()+1 == nextSector; endGarbage = r.garbageTable.End() {
		nextSector = endGarbage.GetFirstSector()
		r.garbageTable.Remove(endGarbage)
	}
	if nextSector != r.nextSector {
		r.nextSector = nextSector
		return r.file.Truncate(int64(r.nextSector) << 12)
	}
	return nil
}

// WriteChunk is a port of RegionLoader::writeChunk: chunkData is zlib-compressed chunk NBT.
func (r *RegionLoader) WriteChunk(x, z int, chunkData []byte) error {
	r.LastUsed = time.Now().Unix()

	length := len(chunkData) + 1
	if length+4 > regionMaxSectorLength {
		return fmt.Errorf("Chunk is too big! %d > %d", length+4, regionMaxSectorLength)
	}
	newSize := int(math.Ceil(float64(length+4) / 4096))
	index, err := getChunkOffset(x, z)
	if err != nil {
		return err
	}

	// look for an unused area big enough to hold this data. this is corruption-resistant (it
	// leaves the old data intact if a failure occurs when writing new data), and also allows the
	// file to become more compact across consecutive writes without introducing a dedicated
	// garbage collection mechanism.
	newLocation := r.garbageTable.Allocate(newSize)
	// if no gaps big enough were found, append to the end of the file instead
	if newLocation == nil {
		newLocation = mustEntry(r.nextSector, newSize, int(time.Now().Unix()))
		r.bumpNextFreeSector(newLocation)
	}

	// write the chunk data into the chosen location
	buf := make([]byte, newSize<<12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(length))
	buf[4] = RegionCompressionZlib
	copy(buf[5:], chunkData)
	if _, err := r.file.WriteAt(buf, int64(newLocation.GetFirstSector())<<12); err != nil {
		return err
	}

	// update the file header - we do this after writing the main data, so that if a failure
	// occurs while writing, the header will still point to the old (intact) copy of the chunk
	oldLocation := r.locationTable[index]
	r.locationTable[index] = newLocation
	if err := r.writeLocationIndex(index); err != nil {
		return err
	}
	if oldLocation != nil {
		return r.disposeGarbageArea(oldLocation)
	}
	return nil
}

// RemoveChunk is a port of RegionLoader::removeChunk.
func (r *RegionLoader) RemoveChunk(x, z int) error {
	index, err := getChunkOffset(x, z)
	if err != nil {
		return err
	}
	oldLocation := r.locationTable[index]
	r.locationTable[index] = nil
	if err := r.writeLocationIndex(index); err != nil {
		return err
	}
	if oldLocation != nil {
		return r.disposeGarbageArea(oldLocation)
	}
	return nil
}

// getChunkOffset is a port of RegionLoader::getChunkOffset.
func getChunkOffset(x, z int) (int, error) {
	if x < 0 || x > 31 || z < 0 || z > 31 {
		return 0, fmt.Errorf("Invalid chunk position in region, expected x/z in range 0-31, got x=%d, z=%d", x, z)
	}
	return x | (z << 5), nil
}

// getChunkCoords is a port of RegionLoader::getChunkCoords.
func getChunkCoords(offset int) (x, z int) {
	return offset & 0x1f, (offset >> 5) & 0x1f
}

// Close is a port of RegionLoader::close.
func (r *RegionLoader) Close() error {
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

// loadLocationTable is a port of RegionLoader::loadLocationTable.
func (r *RegionLoader) loadLocationTable() error {
	headerRaw := make([]byte, regionHeaderLength)
	if n, _ := r.file.ReadAt(headerRaw, 0); n != regionHeaderLength {
		return &CorruptedRegionError{"Corrupted region header (unexpected end of file)"}
	}
	r.locationTable = make([]*RegionLocationTableEntry, 1024)
	for i := 0; i < 1024; i++ {
		index := binary.BigEndian.Uint32(headerRaw[i*4:])
		offset := int(index >> 8)
		sectorCount := int(index & 0xff)
		timestamp := int(binary.BigEndian.Uint32(headerRaw[4096+i*4:]))

		if offset == 0 || sectorCount == 0 {
			r.locationTable[i] = nil
		} else if offset >= RegionFirstSector {
			r.locationTable[i] = mustEntry(offset, sectorCount, timestamp)
			r.bumpNextFreeSector(r.locationTable[i])
		} else {
			x, z := getChunkCoords(i)
			return &CorruptedRegionError{fmt.Sprintf("Invalid region header entry for x=%d z=%d, offset overlaps with header", x, z)}
		}
	}
	if err := r.checkLocationTableValidity(); err != nil {
		return err
	}
	garbage, err := BuildGarbageMapFromLocationTable(r.locationTable)
	if err != nil {
		return err
	}
	r.garbageTable = garbage
	return nil
}

// checkLocationTableValidity is a port of RegionLoader::checkLocationTableValidity.
func (r *RegionLoader) checkLocationTableValidity() error {
	usedOffsets := map[int]int{}
	info, err := r.file.Stat()
	if err != nil {
		return err
	}
	fileSize := info.Size()
	for i := 0; i < 1024; i++ {
		entry := r.locationTable[i]
		if entry == nil {
			continue
		}
		x, z := getChunkCoords(i)
		offset := entry.GetFirstSector()
		fileOffset := int64(offset) << 12

		//TODO: more validity checks
		if fileOffset >= fileSize {
			return &CorruptedRegionError{fmt.Sprintf("Region file location offset x=%d,z=%d points to invalid file location %d", x, z, fileOffset)}
		}
		if existing, ok := usedOffsets[offset]; ok {
			existingX, existingZ := getChunkCoords(existing)
			return &CorruptedRegionError{fmt.Sprintf("Found two chunk offsets (chunk1: x=%d,z=%d, chunk2: x=%d,z=%d) pointing to the file location %d", existingX, existingZ, x, z, fileOffset)}
		}
		usedOffsets[offset] = i
	}
	offsets := make([]int, 0, len(usedOffsets))
	for o := range usedOffsets {
		offsets = append(offsets, o)
	}
	sort.Ints(offsets)
	prevLocationIndex := -1
	for _, o := range offsets {
		locationTableIndex := usedOffsets[o]
		if prevLocationIndex != -1 && r.locationTable[locationTableIndex].Overlaps(r.locationTable[prevLocationIndex]) {
			x, z := getChunkCoords(locationTableIndex)
			prevX, prevZ := getChunkCoords(prevLocationIndex)
			return &CorruptedRegionError{fmt.Sprintf("Overlapping chunks detected in region header (chunk1: x=%d,z=%d, chunk2: x=%d,z=%d)", x, z, prevX, prevZ)}
		}
		prevLocationIndex = locationTableIndex
	}
	return nil
}

// writeLocationIndex is a port of RegionLoader::writeLocationIndex.
func (r *RegionLoader) writeLocationIndex(index int) error {
	entry := r.locationTable[index]
	var location, timestamp uint32
	if entry != nil {
		location = uint32(entry.GetFirstSector()<<8 | entry.GetSectorCount())
		timestamp = uint32(entry.GetTimestamp())
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, location)
	if _, err := r.file.WriteAt(buf, int64(index<<2)); err != nil {
		return err
	}
	binary.BigEndian.PutUint32(buf, timestamp)
	_, err := r.file.WriteAt(buf, int64(4096+(index<<2)))
	return err
}

// createBlank is a port of RegionLoader::createBlank.
func (r *RegionLoader) createBlank() error {
	if err := r.file.Truncate(regionHeaderLength); err != nil { // this fills the file with the null byte
		return err
	}
	r.locationTable = make([]*RegionLocationTableEntry, 1024)
	return nil
}

func (r *RegionLoader) bumpNextFreeSector(entry *RegionLocationTableEntry) {
	r.nextSector = max(r.nextSector, entry.GetLastSector()+1)
}

// GenerateSectorMap is a port of RegionLoader::generateSectorMap (a debugging aid).
func (r *RegionLoader) GenerateSectorMap(usedChar, freeChar string) (string, error) {
	result := []string{}
	for i := 0; i < r.nextSector; i++ {
		result = append(result, freeChar)
	}
	for i := 0; i < RegionFirstSector && i < len(result); i++ {
		result[i] = usedChar
	}
	for _, entry := range r.locationTable {
		if entry == nil {
			continue
		}
		for _, sectorIndex := range entry.GetUsedSectors() {
			if sectorIndex >= len(result) {
				return "", fmt.Errorf("This should never happen...")
			}
			if result[sectorIndex] == usedChar {
				return "", fmt.Errorf("Overlap detected")
			}
			result[sectorIndex] = usedChar
		}
	}
	return strings.Join(result, ""), nil
}

// GetProportionUnusedSpace is a port of RegionLoader::getProportionUnusedSpace.
func (r *RegionLoader) GetProportionUnusedSpace() float64 {
	used := RegionFirstSector // header is always allocated
	for _, entry := range r.locationTable {
		if entry != nil {
			used += entry.GetSectorCount()
		}
	}
	return 1 - float64(used)/float64(r.nextSector)
}

// GetFilePath is a port of RegionLoader::getFilePath.
func (r *RegionLoader) GetFilePath() string { return r.filePath }

// CalculateChunkCount is a port of RegionLoader::calculateChunkCount.
func (r *RegionLoader) CalculateChunkCount() int {
	count := 0
	for i := 0; i < 1024; i++ {
		if r.isChunkGenerated(i) {
			count++
		}
	}
	return count
}
