package itemupgrade

import (
	"fmt"

	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/data/bedrock"
	blockupgrade "pocketmine-go/pocketmine/data/bedrock/block/upgrade"
	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
	"pocketmine-go/pocketmine/nbt"
)

// tagLegacyID is ItemDataUpgrader::TAG_LEGACY_ID: TAG_Short (or TAG_String for Java itemstacks).
const tagLegacyID = "id"

// ItemDataUpgrader is a port of pocketmine\data\bedrock\item\upgrade\ItemDataUpgrader. The block
// state dictionary is the package-level one in data/bedrock (PHP passes the network
// BlockTranslator's BlockStateDictionary, the same data).
type ItemDataUpgrader struct {
	idMetaUpgrader         *ItemIdMetaUpgrader
	legacyIntToStringIdMap *bedrock.LegacyToStringIdMap
	r12ItemIdToBlockIdMap  *R12ItemIdToBlockIdMap
	blockDataUpgrader      *blockupgrade.BlockDataUpgrader
	blockItemIdMap         *bedrockitem.BlockItemIdMap
}

// NewItemDataUpgrader is a port of ItemDataUpgrader::__construct.
func NewItemDataUpgrader(
	idMetaUpgrader *ItemIdMetaUpgrader,
	legacyIntToStringIdMap *bedrock.LegacyToStringIdMap,
	r12ItemIdToBlockIdMap *R12ItemIdToBlockIdMap,
	blockDataUpgrader *blockupgrade.BlockDataUpgrader,
	blockItemIdMap *bedrockitem.BlockItemIdMap,
) *ItemDataUpgrader {
	return &ItemDataUpgrader{
		idMetaUpgrader:         idMetaUpgrader,
		legacyIntToStringIdMap: legacyIntToStringIdMap,
		r12ItemIdToBlockIdMap:  r12ItemIdToBlockIdMap,
		blockDataUpgrader:      blockDataUpgrader,
		blockItemIdMap:         blockItemIdMap,
	}
}

// NewDefaultItemDataUpgrader builds the upgrader GlobalItemDataHandlers::getUpgrader() creates.
func NewDefaultItemDataUpgrader(blockDataUpgrader *blockupgrade.BlockDataUpgrader) (*ItemDataUpgrader, error) {
	schemas, err := LoadSchemas(SchemaFS, "schema/id_meta_upgrade_schema", int(^uint(0)>>1))
	if err != nil {
		return nil, err
	}
	return NewItemDataUpgrader(
		NewItemIdMetaUpgrader(schemas),
		LegacyItemIdToStringIdMap(),
		GetR12ItemIdToBlockIdMap(),
		blockDataUpgrader,
		bedrockitem.GetBlockItemIdMap(),
	), nil
}

func loadingError(format string, args ...any) error {
	return data.NewSavedDataLoadingError(fmt.Sprintf(format, args...))
}

// UpgradeItemTypeDataString is a port of ItemDataUpgrader::upgradeItemTypeDataString (the
// replacement for the legacy ItemFactory::get()).
func (u *ItemDataUpgrader) UpgradeItemTypeDataString(rawNameID string, meta, count int, tag *nbt.CompoundTag) (bedrockitem.SavedItemStackData, error) {
	var blockStateData *bedrock.BlockStateData
	if r12BlockID, ok := u.r12ItemIdToBlockIdMap.ItemIdToBlockId(rawNameID); ok {
		state, err := u.blockDataUpgrader.UpgradeStringIdMeta(r12BlockID, meta)
		if err != nil {
			return bedrockitem.SavedItemStackData{}, loadingError("Failed to deserialize blockstate for legacy blockitem: %v", err)
		}
		blockStateData = &state
	}
	// else: probably a standard item

	newNameID, newMeta := u.idMetaUpgrader.Upgrade(rawNameID, meta)

	//TODO: this won't account for spawn eggs from before 1.16.100 - perhaps we're lucky and they just left the meta in there anyway?
	return bedrockitem.SavedItemStackData{
		TypeData: bedrockitem.SavedItemData{Name: newNameID, Meta: newMeta, Block: blockStateData, Tag: tag},
		Count:    count,
	}, nil
}

// UpgradeItemTypeDataInt is a port of ItemDataUpgrader::upgradeItemTypeDataInt.
func (u *ItemDataUpgrader) UpgradeItemTypeDataInt(legacyNumericID, meta, count int, tag *nbt.CompoundTag) (bedrockitem.SavedItemStackData, error) {
	// do not upgrade the ID beyond this initial step - we need the 1.12 ID for the item ID ->
	// block ID map in the next step
	rawNameID, ok := u.legacyIntToStringIdMap.LegacyToString(legacyNumericID)
	if !ok {
		return bedrockitem.SavedItemStackData{}, loadingError("Unmapped legacy item ID %d", legacyNumericID)
	}
	return u.UpgradeItemTypeDataString(rawNameID, meta, count, tag)
}

// upgradeItemTypeNbt is a port of ItemDataUpgrader::upgradeItemTypeNbt. Returns nil for air.
func (u *ItemDataUpgrader) upgradeItemTypeNbt(tag *nbt.CompoundTag) (*bedrockitem.SavedItemData, error) {
	var rawNameID string
	nameIDTag, _ := tag.GetTag(bedrockitem.SavedItemDataTagName)
	idTag, _ := tag.GetTag(tagLegacyID)
	if name, ok := nameIDTag.(nbt.StringTag); ok {
		// Bedrock 1.6+
		rawNameID = string(name)
	} else if legacyID, ok := idTag.(nbt.ShortTag); ok {
		// Bedrock <= 1.5, PM <= 1.12
		if legacyID == 0 {
			// 0 is a special case for air, which is not a valid item ID. This isn't supposed to be
			// saved, but it appears in some places due to bugs in older versions.
			return nil, nil
		}
		mapped, ok := u.legacyIntToStringIdMap.LegacyToString(int(legacyID))
		if !ok {
			return nil, loadingError("Legacy item ID %d doesn't map to any modern string ID", legacyID)
		}
		rawNameID = mapped
	} else if javaID, ok := idTag.(nbt.StringTag); ok {
		// PC item save format - best we can do here is hope the string IDs match
		rawNameID = string(javaID)
	} else {
		return nil, loadingError("Item stack data should have either a name ID or a legacy ID")
	}

	meta := 0
	if damageTag, ok := tag.GetTag(bedrockitem.SavedItemDataTagDamage); ok {
		damage, isShort := damageTag.(nbt.ShortTag)
		if !isShort {
			return nil, loadingError("Expected a tag of type TAG_Short for \"%s\"", bedrockitem.SavedItemDataTagDamage)
		}
		meta = int(damage)
	}

	var blockStateData *bedrock.BlockStateData
	blockStateNbt, hasBlock, err := tag.GetCompoundTag(bedrockitem.SavedItemDataTagBlock)
	if err != nil {
		return nil, loadingError("%v", err)
	}
	if hasBlock {
		state, err := u.blockDataUpgrader.UpgradeBlockStateNbt(blockStateNbt)
		if err != nil {
			return nil, loadingError("Failed to deserialize blockstate for blockitem: %v", err)
		}
		blockStateData = &state
	} else if r12BlockID, ok := u.r12ItemIdToBlockIdMap.ItemIdToBlockId(rawNameID); ok {
		// this is a legacy blockitem represented by ID + meta
		state, err := u.blockDataUpgrader.UpgradeStringIdMeta(r12BlockID, meta)
		if err != nil {
			return nil, loadingError("Failed to deserialize blockstate for legacy blockitem: %v", err)
		}
		blockStateData = &state
	}
	// else: probably a standard item

	newNameID, newMeta := u.idMetaUpgrader.Upgrade(rawNameID, meta)

	//TODO: Dirty hack to load old skulls from disk: Put this into item upgrade schema's before Mojang makes something with a non 0 default state
	if blockStateData == nil {
		if blockID, ok := u.blockItemIdMap.LookupBlockID(newNameID); ok {
			networkRuntimeID, ok := bedrock.LookupStateIdFromIdMeta(blockID, 0)
			if !ok {
				return nil, loadingError("Failed to find blockstate for blockitem %s", newNameID)
			}
			state, _ := bedrock.GenerateDataFromStateId(networkRuntimeID)
			blockStateData = &state
		}
	}

	//TODO: this won't account for spawn eggs from before 1.16.100 - perhaps we're lucky and they just left the meta in there anyway?
	//TODO: read version from VersionInfo::TAG_WORLD_DATA_VERSION - we may need it to fix up old items
	extra, _, err := tag.GetCompoundTag(bedrockitem.SavedItemDataTagTag)
	if err != nil {
		return nil, loadingError("%v", err)
	}
	return &bedrockitem.SavedItemData{Name: newNameID, Meta: newMeta, Block: blockStateData, Tag: extra}, nil
}

// UpgradeItemStackNbt is a port of ItemDataUpgrader::upgradeItemStackNbt. Returns nil (and no
// error) for air, which older versions of PocketMine-MP saved in some places.
func (u *ItemDataUpgrader) UpgradeItemStackNbt(tag *nbt.CompoundTag) (*bedrockitem.SavedItemStackData, error) {
	savedItemData, err := u.upgradeItemTypeNbt(tag)
	if err != nil || savedItemData == nil {
		return nil, err
	}

	// required
	count, err := tag.GetByte(bedrockitem.SavedItemStackDataTagCount)
	if err != nil {
		return nil, loadingError("%v", err)
	}
	result := &bedrockitem.SavedItemStackData{TypeData: *savedItemData, Count: int(uint8(count))}

	// optional
	if slotTag, ok := tag.GetTag(bedrockitem.SavedItemStackDataTagSlot); ok {
		if slot, ok := slotTag.(nbt.ByteTag); ok {
			v := int(uint8(slot))
			result.Slot = &v
		}
	}
	// PHP: $wasPickedUp !== 0, so an absent tag counts as picked up.
	wasPickedUp := true
	if pickedTag, ok := tag.GetTag(bedrockitem.SavedItemStackDataTagWasPickedUp); ok {
		if v, ok := pickedTag.(nbt.ByteTag); ok {
			wasPickedUp = v != 0
		}
	}
	result.WasPickedUp = &wasPickedUp
	if result.CanPlaceOn, err = stringListTag(tag, bedrockitem.SavedItemStackDataTagCanPlaceOn); err != nil {
		return nil, err
	}
	if result.CanDestroy, err = stringListTag(tag, bedrockitem.SavedItemStackDataTagCanDestroy); err != nil {
		return nil, err
	}
	return result, nil
}

// stringListTag is CompoundTag::getListTag($name, StringTag::class) mapped to its values.
func stringListTag(tag *nbt.CompoundTag, name string) ([]string, error) {
	list, ok, err := tag.GetListTag(name)
	if err != nil {
		return nil, loadingError("%v", err)
	}
	if !ok {
		return nil, nil
	}
	var result []string
	for _, t := range list.Values() {
		s, isString := t.(nbt.StringTag)
		if !isString {
			return nil, loadingError("Expected all tags of %s to be TAG_String", name)
		}
		result = append(result, string(s))
	}
	return result, nil
}

// GetIdMetaUpgrader is a port of ItemDataUpgrader::getIdMetaUpgrader.
func (u *ItemDataUpgrader) GetIdMetaUpgrader() *ItemIdMetaUpgrader { return u.idMetaUpgrader }
