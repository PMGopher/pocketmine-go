package itemupgrade

import (
	"embed"
	"sync"

	"pocketmine-go/pocketmine/data/bedrock"
)

// SchemaFS holds the vendored pmmp/BedrockItemUpgradeSchema data (PHP's
// BEDROCK_ITEM_UPGRADE_SCHEMA_PATH).
//
//go:embed schema/id_meta_upgrade_schema/*.json schema/item_legacy_id_map.json schema/1.12.0_item_id_to_block_id_map.json
var SchemaFS embed.FS

var (
	legacyItemIDMapOnce sync.Once
	legacyItemIDMap     *bedrock.LegacyToStringIdMap
)

// LegacyItemIdToStringIdMap is a port of LegacyItemIdToStringIdMap::getInstance():
// item_legacy_id_map.json.
func LegacyItemIdToStringIdMap() *bedrock.LegacyToStringIdMap {
	legacyItemIDMapOnce.Do(func() {
		data, err := SchemaFS.ReadFile("schema/item_legacy_id_map.json")
		if err != nil {
			panic(err)
		}
		legacyItemIDMap = bedrock.NewLegacyToStringIdMap(data)
	})
	return legacyItemIDMap
}
