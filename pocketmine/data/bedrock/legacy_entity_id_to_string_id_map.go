package bedrock

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed assets/entity_id_map.json
var entityIDMapJSON []byte

// LegacyToStringIdMap is a port of pocketmine\data\bedrock\LegacyToStringIdMap: legacy numeric IDs
// to their modern string IDs.
type LegacyToStringIdMap struct {
	legacyToString map[int]string
}

func newLegacyToStringIdMap(data []byte) *LegacyToStringIdMap {
	var stringToLegacy map[string]int
	if err := json.Unmarshal(data, &stringToLegacy); err != nil {
		panic(fmt.Sprintf("bedrock: invalid format of ID map: %v", err))
	}
	m := &LegacyToStringIdMap{legacyToString: make(map[int]string, len(stringToLegacy))}
	for stringID, legacyID := range stringToLegacy {
		m.legacyToString[legacyID] = stringID
	}
	return m
}

// LegacyToString returns the string ID for legacy, if known.
func (m *LegacyToStringIdMap) LegacyToString(legacy int) (string, bool) {
	s, ok := m.legacyToString[legacy]
	return s, ok
}

// GetLegacyToStringMap returns the whole mapping.
func (m *LegacyToStringIdMap) GetLegacyToStringMap() map[int]string { return m.legacyToString }

// Add is a port of LegacyToStringIdMap::add. Panics if legacy is already mapped to a different
// string, like PHP's InvalidArgumentException.
func (m *LegacyToStringIdMap) Add(s string, legacy int) {
	if existing, ok := m.legacyToString[legacy]; ok {
		if existing == s {
			return
		}
		panic(fmt.Sprintf("Legacy ID %d is already mapped to string %s", legacy, existing))
	}
	m.legacyToString[legacy] = s
}

var (
	legacyEntityIDMapOnce sync.Once
	legacyEntityIDMap     *LegacyToStringIdMap
)

// LegacyEntityIdToStringIdMap is the port of LegacyEntityIdToStringIdMap::getInstance(), backed by
// BedrockData's entity_id_map.json (vendored from pmmp/BedrockData at the same revision as the
// other assets in this package).
func LegacyEntityIdToStringIdMap() *LegacyToStringIdMap {
	legacyEntityIDMapOnce.Do(func() {
		legacyEntityIDMap = newLegacyToStringIdMap(entityIDMapJSON)
	})
	return legacyEntityIDMap
}
