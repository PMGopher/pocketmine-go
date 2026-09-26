// Package itemupgrade is a port of pocketmine\data\bedrock\item\upgrade: upgrading items saved by
// older Minecraft/PocketMine-MP versions (legacy numeric IDs, renamed IDs and meta-based
// variants) to current item data, using pmmp's BedrockItemUpgradeSchema data (vendored in
// schema/).
package itemupgrade

import "strings"

// ItemIdMetaUpgradeSchema is a port of pocketmine\data\bedrock\item\upgrade\ItemIdMetaUpgradeSchema.
type ItemIdMetaUpgradeSchema struct {
	renamedIds    map[string]string
	remappedMetas map[string]map[int]string
	schemaID      int
}

// NewItemIdMetaUpgradeSchema is a port of ItemIdMetaUpgradeSchema::__construct.
func NewItemIdMetaUpgradeSchema(renamedIds map[string]string, remappedMetas map[string]map[int]string, schemaID int) *ItemIdMetaUpgradeSchema {
	return &ItemIdMetaUpgradeSchema{renamedIds: renamedIds, remappedMetas: remappedMetas, schemaID: schemaID}
}

func (s *ItemIdMetaUpgradeSchema) GetSchemaID() int                 { return s.schemaID }
func (s *ItemIdMetaUpgradeSchema) GetRenamedIds() map[string]string { return s.renamedIds }
func (s *ItemIdMetaUpgradeSchema) GetRemappedMetas() map[string]map[int]string {
	return s.remappedMetas
}

// RenameID is a port of ItemIdMetaUpgradeSchema::renameId.
func (s *ItemIdMetaUpgradeSchema) RenameID(id string) (string, bool) {
	newID, ok := s.renamedIds[strings.ToLower(id)]
	return newID, ok
}

// RemapMeta is a port of ItemIdMetaUpgradeSchema::remapMeta.
func (s *ItemIdMetaUpgradeSchema) RemapMeta(id string, meta int) (string, bool) {
	newID, ok := s.remappedMetas[strings.ToLower(id)][meta]
	return newID, ok
}
