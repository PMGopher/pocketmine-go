// Package region is a port of pocketmine\world\format\io\region: the read-only Anvil, McRegion
// and PMAnvil world formats (Java Edition and old PocketMine-MP worlds), which WorldManager
// converts to LevelDB when loading them.
package region

import "fmt"

// RegionLocationTableEntry is a port of RegionLocationTableEntry: where a chunk's data lives in a
// region file, in 4KiB sectors.
type RegionLocationTableEntry struct {
	firstSector int
	sectorCount int
	timestamp   int
}

// NewRegionLocationTableEntry is a port of RegionLocationTableEntry::__construct.
func NewRegionLocationTableEntry(firstSector, sectorCount, timestamp int) (*RegionLocationTableEntry, error) {
	if firstSector < 0 || firstSector >= 1<<24 {
		return nil, fmt.Errorf("Start sector must be positive, got %d", firstSector)
	}
	if sectorCount < 1 {
		return nil, fmt.Errorf("Sector count must be positive, got %d", sectorCount)
	}
	return &RegionLocationTableEntry{firstSector: firstSector, sectorCount: sectorCount, timestamp: timestamp}, nil
}

func mustEntry(firstSector, sectorCount, timestamp int) *RegionLocationTableEntry {
	e, err := NewRegionLocationTableEntry(firstSector, sectorCount, timestamp)
	if err != nil {
		panic(err)
	}
	return e
}

func (e *RegionLocationTableEntry) GetFirstSector() int { return e.firstSector }
func (e *RegionLocationTableEntry) GetLastSector() int  { return e.firstSector + e.sectorCount - 1 }
func (e *RegionLocationTableEntry) GetSectorCount() int { return e.sectorCount }
func (e *RegionLocationTableEntry) GetTimestamp() int   { return e.timestamp }

// GetUsedSectors is a port of RegionLocationTableEntry::getUsedSectors.
func (e *RegionLocationTableEntry) GetUsedSectors() []int {
	sectors := make([]int, 0, e.sectorCount)
	for i := e.GetFirstSector(); i <= e.GetLastSector(); i++ {
		sectors = append(sectors, i)
	}
	return sectors
}

// Overlaps is a port of RegionLocationTableEntry::overlaps.
func (e *RegionLocationTableEntry) Overlaps(other *RegionLocationTableEntry) bool {
	overlapCheck := func(entry1, entry2 *RegionLocationTableEntry) bool {
		entry1Last := entry1.GetLastSector()
		entry2Last := entry2.GetLastSector()
		return (entry2.firstSector >= entry1.firstSector && entry2.firstSector <= entry1Last) ||
			(entry2Last >= entry1.firstSector && entry2Last <= entry1Last)
	}
	return overlapCheck(e, other) || overlapCheck(other, e)
}
