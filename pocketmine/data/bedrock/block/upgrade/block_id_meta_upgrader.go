package blockupgrade

import (
	"fmt"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
)

// BlockIdMetaUpgrader is a port of pocketmine\data\bedrock\block\upgrade\BlockIdMetaUpgrader:
// legacy (string or numeric ID + meta) blocks to modern blockstates.
type BlockIdMetaUpgrader struct {
	mappingTable       map[string]map[int]bedrock.BlockStateData
	legacyNumericIdMap *bedrock.LegacyToStringIdMap
}

// NewBlockIdMetaUpgrader is a port of BlockIdMetaUpgrader::__construct.
func NewBlockIdMetaUpgrader(mappingTable map[string]map[int]bedrock.BlockStateData, legacyNumericIdMap *bedrock.LegacyToStringIdMap) *BlockIdMetaUpgrader {
	return &BlockIdMetaUpgrader{mappingTable: mappingTable, legacyNumericIdMap: legacyNumericIdMap}
}

// FromStringIdMeta is a port of BlockIdMetaUpgrader::fromStringIdMeta.
func (u *BlockIdMetaUpgrader) FromStringIdMeta(id string, meta int) (bedrock.BlockStateData, error) {
	if metas, ok := u.mappingTable[id]; ok {
		if state, ok := metas[meta]; ok {
			return state, nil
		}
		if state, ok := metas[0]; ok {
			return state, nil
		}
	}
	return bedrock.BlockStateData{}, &bedrock.BlockStateDeserializeError{Message: "Unknown legacy block string ID " + id}
}

// FromIntIdMeta is a port of BlockIdMetaUpgrader::fromIntIdMeta.
func (u *BlockIdMetaUpgrader) FromIntIdMeta(id, meta int) (bedrock.BlockStateData, error) {
	stringID, ok := u.legacyNumericIdMap.LegacyToString(id)
	if !ok {
		return bedrock.BlockStateData{}, &bedrock.BlockStateDeserializeError{Message: fmt.Sprintf("Unknown legacy block numeric ID %d", id)}
	}
	return u.FromStringIdMeta(stringID, meta)
}

// AddIntIdToStringIdMapping is a port of BlockIdMetaUpgrader::addIntIdToStringIdMapping.
func (u *BlockIdMetaUpgrader) AddIntIdToStringIdMapping(intID int, stringID string) {
	u.legacyNumericIdMap.Add(stringID, intID)
}

// AddIdMetaToStateMapping is a port of BlockIdMetaUpgrader::addIdMetaToStateMapping. Panics if a
// mapping for that ID and meta already exists.
func (u *BlockIdMetaUpgrader) AddIdMetaToStateMapping(stringID string, meta int, stateData bedrock.BlockStateData) {
	if _, ok := u.mappingTable[stringID][meta]; ok {
		panic(fmt.Sprintf("A mapping for %s:%d already exists", stringID, meta))
	}
	if u.mappingTable[stringID] == nil {
		u.mappingTable[stringID] = map[int]bedrock.BlockStateData{}
	}
	u.mappingTable[stringID][meta] = stateData
}

// LoadBlockIdMetaUpgraderFromString is a port of BlockIdMetaUpgrader::loadFromString: the
// id_meta_to_nbt table (unsigned varint counts, varint-prefixed IDs, little-endian NBT states),
// with every state upgraded to the current version.
func LoadBlockIdMetaUpgraderFromString(data []byte, idMap *bedrock.LegacyToStringIdMap, blockStateUpgrader *BlockStateUpgrader) (*BlockIdMetaUpgrader, error) {
	mappingTable := map[string]map[int]bedrock.BlockStateData{}
	offset := 0
	nbtReader := nbt.NewLittleEndianSerializer()

	idCount, err := binaryutils.ReadUnsignedVarInt(data, &offset)
	if err != nil {
		return nil, err
	}
	for idIndex := uint32(0); idIndex < idCount; idIndex++ {
		idLen, err := binaryutils.ReadUnsignedVarInt(data, &offset)
		if err != nil {
			return nil, err
		}
		if offset+int(idLen) > len(data) {
			return nil, fmt.Errorf("unexpected end of legacy state map data")
		}
		id := string(data[offset : offset+int(idLen)])
		offset += int(idLen)

		metaCount, err := binaryutils.ReadUnsignedVarInt(data, &offset)
		if err != nil {
			return nil, err
		}
		for metaIndex := uint32(0); metaIndex < metaCount; metaIndex++ {
			meta, err := binaryutils.ReadUnsignedVarInt(data, &offset)
			if err != nil {
				return nil, err
			}
			root, newOffset, err := nbtReader.Read(data, offset, 512)
			if err != nil {
				return nil, err
			}
			offset = newOffset
			tag, err := root.MustGetCompoundTag()
			if err != nil {
				return nil, err
			}
			state, err := bedrock.BlockStateDataFromNbt(tag)
			if err != nil {
				return nil, err
			}
			if mappingTable[id] == nil {
				mappingTable[id] = map[int]bedrock.BlockStateData{}
			}
			mappingTable[id][int(meta)] = blockStateUpgrader.Upgrade(state)
		}
	}
	if offset < len(data) {
		return nil, fmt.Errorf("Unexpected trailing data in legacy state map data")
	}
	return NewBlockIdMetaUpgrader(mappingTable, idMap), nil
}
