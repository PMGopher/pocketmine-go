// Package model is a port of pocketmine\data\bedrock\item\upgrade\model.
package model

// ItemIdMetaUpgradeSchemaModel is a port of ItemIdMetaUpgradeSchemaModel: the JSON shape of a
// BedrockItemUpgradeSchema id_meta_upgrade_schema file.
type ItemIdMetaUpgradeSchemaModel struct {
	RenamedIds    map[string]string            `json:"renamedIds"`
	RemappedMetas map[string]map[string]string `json:"remappedMetas"`
}
