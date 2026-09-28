package blockupgrade

import (
	"embed"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
)

// SchemaFS holds the vendored pmmp/BedrockBlockUpgradeSchema data (PHP's
// BEDROCK_BLOCK_UPGRADE_SCHEMA_PATH).
//
//go:embed schema/nbt_upgrade_schema/*.json schema/0351_1.26.40_to_1.26.50.json schema/id_meta_to_nbt/1.12.0.bin schema/block_legacy_id_map.json
var SchemaFS embed.FS

// BlockDataUpgrader is a port of pocketmine\data\bedrock\block\upgrade\BlockDataUpgrader.
type BlockDataUpgrader struct {
	blockIdMetaUpgrader *BlockIdMetaUpgrader
	blockStateUpgrader  *BlockStateUpgrader
}

// NewBlockDataUpgrader is a port of BlockDataUpgrader::__construct.
func NewBlockDataUpgrader(blockIdMetaUpgrader *BlockIdMetaUpgrader, blockStateUpgrader *BlockStateUpgrader) *BlockDataUpgrader {
	return &BlockDataUpgrader{blockIdMetaUpgrader: blockIdMetaUpgrader, blockStateUpgrader: blockStateUpgrader}
}

// UpgradeIntIdMeta is a port of BlockDataUpgrader::upgradeIntIdMeta.
func (u *BlockDataUpgrader) UpgradeIntIdMeta(id, meta int) (bedrock.BlockStateData, error) {
	return u.blockIdMetaUpgrader.FromIntIdMeta(id, meta)
}

// UpgradeStringIdMeta is a port of BlockDataUpgrader::upgradeStringIdMeta.
func (u *BlockDataUpgrader) UpgradeStringIdMeta(id string, meta int) (bedrock.BlockStateData, error) {
	return u.blockIdMetaUpgrader.FromStringIdMeta(id, meta)
}

// UpgradeBlockStateNbt is a port of BlockDataUpgrader::upgradeBlockStateNbt.
func (u *BlockDataUpgrader) UpgradeBlockStateNbt(tag *nbt.CompoundTag) (bedrock.BlockStateData, error) {
	var data bedrock.BlockStateData
	_, hasName := tag.GetTag("name")
	_, hasVal := tag.GetTag("val")
	if hasName && hasVal {
		// Legacy (pre-1.13) blockstate - upgrade it to a version we understand
		id, err := tag.GetString("name")
		if err != nil {
			return bedrock.BlockStateData{}, &bedrock.BlockStateDeserializeError{Message: err.Error()}
		}
		meta, err := tag.GetShort("val")
		if err != nil {
			return bedrock.BlockStateData{}, &bedrock.BlockStateDeserializeError{Message: err.Error()}
		}
		if data, err = u.UpgradeStringIdMeta(string(id), int(meta)); err != nil {
			return bedrock.BlockStateData{}, err
		}
	} else {
		// Modern (post-1.13) blockstate
		var err error
		if data, err = bedrock.BlockStateDataFromNbt(tag); err != nil {
			return bedrock.BlockStateData{}, err
		}
	}
	return u.blockStateUpgrader.Upgrade(data), nil
}

// GetBlockStateUpgrader is a port of BlockDataUpgrader::getBlockStateUpgrader.
func (u *BlockDataUpgrader) GetBlockStateUpgrader() *BlockStateUpgrader { return u.blockStateUpgrader }

// GetBlockIdMetaUpgrader is a port of BlockDataUpgrader::getBlockIdMetaUpgrader.
func (u *BlockDataUpgrader) GetBlockIdMetaUpgrader() *BlockIdMetaUpgrader {
	return u.blockIdMetaUpgrader
}

// NewDefaultBlockDataUpgrader builds the upgrader GlobalBlockStateHandlers::getUpgrader() creates:
// every nbt_upgrade_schema file and the 1.12.0 id_meta_to_nbt table.
func NewDefaultBlockDataUpgrader() (*BlockDataUpgrader, error) {
	schemas, err := LoadSchemas(SchemaFS, "schema/nbt_upgrade_schema", int(^uint(0)>>1))
	if err != nil {
		return nil, err
	}
	blockStateUpgrader := NewBlockStateUpgrader(schemas)
	if err := addBedrock12650Schema(blockStateUpgrader); err != nil {
		return nil, err
	}
	table, err := SchemaFS.ReadFile("schema/id_meta_to_nbt/1.12.0.bin")
	if err != nil {
		return nil, err
	}
	idMeta, err := LoadBlockIdMetaUpgraderFromString(table, LegacyBlockIdToStringIdMap(), blockStateUpgrader)
	if err != nil {
		return nil, err
	}
	return NewBlockDataUpgrader(idMeta, blockStateUpgrader), nil
}

// blockStateVersion12650 is Bedrock 1.26.50.0 as a block state version (major<<24 | minor<<16 |
// patch<<8 | revision).
const blockStateVersion12650 = 1<<24 | 26<<16 | 50<<8

// addBedrock12650Schema adds pmmp's 0351_1.26.40_to_1.26.50 schema, which PocketMine-MP 5.44.4
// doesn't load (it ships at the root of BedrockBlockUpgradeSchema, outside nbt_upgrade_schema/,
// because 5.44.4 only supports clients up to 1.26.30). This server speaks 1.26.50, whose stairs,
// fences, panes, bars and tripwire have new properties (minecraft:corner, minecraft:connection_*);
// without the schema, those blocks in worlds saved by older versions became "update!" blocks.
//
// Like every recent pmmp schema it's stamped 1.21.60.33, but worlds saved by vanilla 1.26.30 stamp
// their states 1.26.30, newer than that, so BlockStateUpgrader::upgrade would skip it. It's added as
// the 1.26.50 upgrade it is instead. Its only changes are added properties, which are only set when
// missing, so states that already have them are unchanged. The output version stays the palette's.
func addBedrock12650Schema(upgrader *BlockStateUpgrader) error {
	raw, err := SchemaFS.ReadFile("schema/0351_1.26.40_to_1.26.50.json")
	if err != nil {
		return err
	}
	schema, err := LoadSchemaFromString(raw, 351)
	if err != nil {
		return err
	}
	schema.versionID = blockStateVersion12650
	outputVersion := upgrader.outputVersion
	upgrader.AddSchema(schema)
	upgrader.outputVersion = outputVersion
	return nil
}
