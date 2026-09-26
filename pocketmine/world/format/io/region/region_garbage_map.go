package region

import (
	"fmt"
	"sort"
	"time"
)

// RegionGarbageMap is a port of RegionGarbageMap: the unused areas of a region file, keyed by first
// sector.
type RegionGarbageMap struct {
	entries map[int]*RegionLocationTableEntry
	clean   bool
}

// NewRegionGarbageMap is a port of RegionGarbageMap::__construct.
func NewRegionGarbageMap(entries []*RegionLocationTableEntry) *RegionGarbageMap {
	m := &RegionGarbageMap{entries: map[int]*RegionLocationTableEntry{}}
	for _, e := range entries {
		m.entries[e.GetFirstSector()] = e
	}
	return m
}

// BuildGarbageMapFromLocationTable is a port of RegionGarbageMap::buildFromLocationTable: the gaps
// between the used areas.
func BuildGarbageMapFromLocationTable(locationTable []*RegionLocationTableEntry) (*RegionGarbageMap, error) {
	usedMap := map[int]*RegionLocationTableEntry{}
	for _, e := range locationTable {
		if e == nil {
			continue
		}
		if _, ok := usedMap[e.GetFirstSector()]; ok {
			return nil, fmt.Errorf("Overlapping entries detected")
		}
		usedMap[e.GetFirstSector()] = e
	}
	keys := make([]int, 0, len(usedMap))
	for k := range usedMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var garbage []*RegionLocationTableEntry
	var prevEntry *RegionLocationTableEntry
	for _, k := range keys {
		entry := usedMap[k]
		prevEndPlusOne := RegionFirstSector
		if prevEntry != nil {
			prevEndPlusOne = prevEntry.GetLastSector() + 1
		}
		currentStart := entry.GetFirstSector()
		if prevEndPlusOne < currentStart {
			// found a gap in the table
			garbage = append(garbage, mustEntry(prevEndPlusOne, currentStart-prevEndPlusOne, 0))
		} else if prevEndPlusOne > currentStart {
			// current entry starts inside the previous. This would be a bug since RegionLoader should prevent this
			return nil, fmt.Errorf("Overlapping entries detected")
		}
		prevEntry = entry
	}
	return NewRegionGarbageMap(garbage), nil
}

// sorted returns getArray()'s entries in first-sector order, merging adjacent ones.
func (m *RegionGarbageMap) sorted() []*RegionLocationTableEntry {
	keys := make([]int, 0, len(m.entries))
	for k := range m.entries {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if !m.clean {
		prevIndex := -1
		for _, k := range keys {
			entry := m.entries[k]
			if prevIndex != -1 && m.entries[prevIndex].GetLastSector()+1 == entry.GetFirstSector() {
				prev := m.entries[prevIndex]
				m.entries[prevIndex] = mustEntry(prev.GetFirstSector(), prev.GetSectorCount()+entry.GetSectorCount(), 0)
				delete(m.entries, k)
			} else {
				prevIndex = k
			}
		}
		m.clean = true
		keys = keys[:0]
		for k := range m.entries {
			keys = append(keys, k)
		}
		sort.Ints(keys)
	}
	result := make([]*RegionLocationTableEntry, len(keys))
	for i, k := range keys {
		result[i] = m.entries[k]
	}
	return result
}

// GetArray is a port of RegionGarbageMap::getArray: the entries, adjacent ones merged.
func (m *RegionGarbageMap) GetArray() []*RegionLocationTableEntry { return m.sorted() }

// Add is a port of RegionGarbageMap::add.
func (m *RegionGarbageMap) Add(entry *RegionLocationTableEntry) error {
	if _, ok := m.entries[entry.GetFirstSector()]; ok {
		return fmt.Errorf("Overlapping entry starting at %d", entry.GetFirstSector())
	}
	m.entries[entry.GetFirstSector()] = entry
	m.clean = false
	return nil
}

// Remove is a port of RegionGarbageMap::remove.
func (m *RegionGarbageMap) Remove(entry *RegionLocationTableEntry) {
	// removal doesn't affect ordering and shouldn't affect fragmentation
	delete(m.entries, entry.GetFirstSector())
}

// End is a port of RegionGarbageMap::end: the last entry, or nil.
func (m *RegionGarbageMap) End() *RegionLocationTableEntry {
	entries := m.sorted()
	if len(entries) == 0 {
		return nil
	}
	return entries[len(entries)-1]
}

// Allocate is a port of RegionGarbageMap::allocate: takes newSize sectors from the first unused
// area big enough, or returns nil.
func (m *RegionGarbageMap) Allocate(newSize int) *RegionLocationTableEntry {
	for _, candidate := range m.sorted() {
		candidateSize := candidate.GetSectorCount()
		if candidateSize < newSize {
			continue
		}
		newLocation := mustEntry(candidate.GetFirstSector(), newSize, int(time.Now().Unix()))
		m.Remove(candidate)
		if candidateSize > newSize { // we're not using the whole area, just take part of it
			_ = m.Add(mustEntry(candidate.GetFirstSector()+newSize, candidateSize-newSize, 0))
		}
		return newLocation
	}
	return nil
}
