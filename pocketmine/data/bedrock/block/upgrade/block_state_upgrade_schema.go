// Package blockupgrade is a port of pocketmine\data\bedrock\block\upgrade: upgrading blockstates
// saved by older Minecraft/PocketMine-MP versions (legacy numeric IDs, legacy name+meta states and
// older name+properties states) to the current blockstate format, using pmmp's
// BedrockBlockUpgradeSchema data (vendored in schema/).
//
// Block state values are represented like bedrock.BlockStateData.States: int32 (TAG_Int), uint8
// (TAG_Byte) and string (TAG_String).
package blockupgrade

import "sort"

// BlockStateUpgradeSchemaValueRemap is a port of BlockStateUpgradeSchemaValueRemap.
type BlockStateUpgradeSchemaValueRemap struct {
	Old any
	New any
}

// FlattenedPropertyType names the tag type a flattened property must have (PHP stores the tag
// class name).
type FlattenedPropertyType int

const (
	FlattenedTypeString FlattenedPropertyType = iota
	FlattenedTypeByte
	FlattenedTypeInt
)

// matches reports whether v is a value of this tag type.
func (t FlattenedPropertyType) matches(v any) bool {
	switch t {
	case FlattenedTypeString:
		_, ok := v.(string)
		return ok
	case FlattenedTypeByte:
		_, ok := v.(uint8)
		return ok
	case FlattenedTypeInt:
		_, ok := v.(int32)
		return ok
	}
	return false
}

// BlockStateUpgradeSchemaFlattenInfo is a port of BlockStateUpgradeSchemaFlattenInfo.
type BlockStateUpgradeSchemaFlattenInfo struct {
	Prefix                string
	FlattenedProperty     string
	Suffix                string
	FlattenedValueRemaps  map[string]string
	FlattenedPropertyType FlattenedPropertyType
}

// Equals is a port of BlockStateUpgradeSchemaFlattenInfo::equals.
func (f *BlockStateUpgradeSchemaFlattenInfo) Equals(that *BlockStateUpgradeSchemaFlattenInfo) bool {
	if f.Prefix != that.Prefix || f.FlattenedProperty != that.FlattenedProperty || f.Suffix != that.Suffix ||
		f.FlattenedPropertyType != that.FlattenedPropertyType || len(f.FlattenedValueRemaps) != len(that.FlattenedValueRemaps) {
		return false
	}
	for k, v := range f.FlattenedValueRemaps {
		if w, ok := that.FlattenedValueRemaps[k]; !ok || w != v {
			return false
		}
	}
	return true
}

// BlockStateUpgradeSchemaBlockRemap is a port of BlockStateUpgradeSchemaBlockRemap. Exactly one of
// NewName and NewFlattenedName is set (PHP: string|BlockStateUpgradeSchemaFlattenInfo).
type BlockStateUpgradeSchemaBlockRemap struct {
	OldState         map[string]any
	NewName          string
	NewFlattenedName *BlockStateUpgradeSchemaFlattenInfo
	NewState         map[string]any
	CopiedState      []string
}

// BlockStateUpgradeSchema is a port of BlockStateUpgradeSchema.
type BlockStateUpgradeSchema struct {
	RenamedIds             map[string]string
	AddedProperties        map[string]map[string]any
	RemovedProperties      map[string][]string
	RenamedProperties      map[string]map[string]string
	RemappedPropertyValues map[string]map[string][]BlockStateUpgradeSchemaValueRemap
	FlattenedProperties    map[string]*BlockStateUpgradeSchemaFlattenInfo
	RemappedStates         map[string][]BlockStateUpgradeSchemaBlockRemap

	MaxVersionMajor, MaxVersionMinor, MaxVersionPatch, MaxVersionRevision int

	versionID int32
	schemaID  int
}

// NewBlockStateUpgradeSchema is a port of BlockStateUpgradeSchema::__construct.
func NewBlockStateUpgradeSchema(major, minor, patch, revision, schemaID int) *BlockStateUpgradeSchema {
	return &BlockStateUpgradeSchema{
		RenamedIds:             map[string]string{},
		AddedProperties:        map[string]map[string]any{},
		RemovedProperties:      map[string][]string{},
		RenamedProperties:      map[string]map[string]string{},
		RemappedPropertyValues: map[string]map[string][]BlockStateUpgradeSchemaValueRemap{},
		FlattenedProperties:    map[string]*BlockStateUpgradeSchemaFlattenInfo{},
		RemappedStates:         map[string][]BlockStateUpgradeSchemaBlockRemap{},
		MaxVersionMajor:        major,
		MaxVersionMinor:        minor,
		MaxVersionPatch:        patch,
		MaxVersionRevision:     revision,
		versionID:              int32(major<<24 | minor<<16 | patch<<8 | revision),
		schemaID:               schemaID,
	}
}

// GetVersionID is a port of BlockStateUpgradeSchema::getVersionId.
func (s *BlockStateUpgradeSchema) GetVersionID() int32 { return s.versionID }

// GetSchemaID is a port of BlockStateUpgradeSchema::getSchemaId.
func (s *BlockStateUpgradeSchema) GetSchemaID() int { return s.schemaID }

// IsEmpty is a port of BlockStateUpgradeSchema::isEmpty.
func (s *BlockStateUpgradeSchema) IsEmpty() bool {
	return len(s.RenamedIds) == 0 && len(s.AddedProperties) == 0 && len(s.RemovedProperties) == 0 &&
		len(s.RenamedProperties) == 0 && len(s.RemappedPropertyValues) == 0 &&
		len(s.FlattenedProperties) == 0 && len(s.RemappedStates) == 0
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
