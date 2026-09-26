package blockupgrade

import (
	"sync"

	"pocketmine-go/pocketmine/data/bedrock"
)

var (
	legacyBlockIDMapOnce sync.Once
	legacyBlockIDMap     *bedrock.LegacyToStringIdMap
)

// LegacyBlockIdToStringIdMap is a port of LegacyBlockIdToStringIdMap::getInstance():
// block_legacy_id_map.json.
func LegacyBlockIdToStringIdMap() *bedrock.LegacyToStringIdMap {
	legacyBlockIDMapOnce.Do(func() {
		data, err := SchemaFS.ReadFile("schema/block_legacy_id_map.json")
		if err != nil {
			panic(err)
		}
		legacyBlockIDMap = bedrock.NewLegacyToStringIdMap(data)
	})
	return legacyBlockIDMap
}
