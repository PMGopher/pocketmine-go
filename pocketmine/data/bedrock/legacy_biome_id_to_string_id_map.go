package bedrock

import "sync"

// LegacyBiomeIdToStringIdMap is a port of pocketmine\data\bedrock\LegacyBiomeIdToStringIdMap:
// numeric biome IDs to their string IDs. PHP loads it from BedrockData's biome_id_map.json, which
// has no 1.26.50 release; this builds it from the 1.26.50 BiomeDefinitionList this server sends,
// which lists every biome the client knows by ID and name.
type LegacyBiomeIdToStringIdMap struct {
	legacyToString map[int]string
	stringToLegacy map[string]int
}

var (
	legacyBiomeIDMapOnce sync.Once
	legacyBiomeIDMap     *LegacyBiomeIdToStringIdMap
)

// GetLegacyBiomeIdToStringIdMap is LegacyBiomeIdToStringIdMap::getInstance.
func GetLegacyBiomeIdToStringIdMap() *LegacyBiomeIdToStringIdMap {
	legacyBiomeIDMapOnce.Do(func() {
		m := &LegacyBiomeIdToStringIdMap{legacyToString: map[int]string{}, stringToLegacy: map[string]int{}}
		pk := BiomeDefinitionList()
		for _, def := range pk.BiomeDefinitions {
			if int(def.NameIndex) < 0 || int(def.NameIndex) >= len(pk.StringList) {
				continue
			}
			name := pk.StringList[def.NameIndex]
			id := int(def.BiomeID)
			if id < 0 {
				vanillaID, ok := vanillaBiomeIDs[name]
				if !ok {
					continue
				}
				id = vanillaID
			}
			m.legacyToString[id] = name
			m.stringToLegacy[name] = id
		}
		legacyBiomeIDMap = m
	})
	return legacyBiomeIDMap
}

// LegacyToString is a port of LegacyToStringIdMap::legacyToString (false for an unknown ID).
func (m *LegacyBiomeIdToStringIdMap) LegacyToString(legacy int) (string, bool) {
	s, ok := m.legacyToString[legacy]
	return s, ok
}

// StringToLegacy is a port of LegacyToStringIdMap::stringToLegacy (false for an unknown name).
func (m *LegacyBiomeIdToStringIdMap) StringToLegacy(s string) (int, bool) {
	id, ok := m.stringToLegacy[s]
	return id, ok
}
