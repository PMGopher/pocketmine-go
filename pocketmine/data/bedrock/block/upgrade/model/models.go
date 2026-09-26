// Package model is a port of pocketmine\data\bedrock\block\upgrade\model: the JSON shape of a
// BedrockBlockUpgradeSchema file. Optional fields are pointers/nil maps so that "absent" and
// "empty" can be told apart, like PHP's uninitialized typed properties.
package model

// BlockStateUpgradeSchemaModelTag is a port of BlockStateUpgradeSchemaModelTag: exactly one of the
// three fields must be set.
type BlockStateUpgradeSchemaModelTag struct {
	Byte   *int    `json:"byte,omitempty"`
	Int    *int    `json:"int,omitempty"`
	String *string `json:"string,omitempty"`
}

// BlockStateUpgradeSchemaModelValueRemap is a port of BlockStateUpgradeSchemaModelValueRemap.
type BlockStateUpgradeSchemaModelValueRemap struct {
	Old BlockStateUpgradeSchemaModelTag `json:"old"`
	New BlockStateUpgradeSchemaModelTag `json:"new"`
}

// BlockStateUpgradeSchemaModelFlattenInfo is a port of BlockStateUpgradeSchemaModelFlattenInfo.
type BlockStateUpgradeSchemaModelFlattenInfo struct {
	Prefix                string            `json:"prefix"`
	FlattenedProperty     string            `json:"flattenedProperty"`
	FlattenedPropertyType *string           `json:"flattenedPropertyType,omitempty"`
	Suffix                string            `json:"suffix"`
	FlattenedValueRemaps  map[string]string `json:"flattenedValueRemaps,omitempty"`
}

// BlockStateUpgradeSchemaModelBlockRemap is a port of BlockStateUpgradeSchemaModelBlockRemap:
// exactly one of NewName and NewFlattenedName must be set.
type BlockStateUpgradeSchemaModelBlockRemap struct {
	OldState         map[string]BlockStateUpgradeSchemaModelTag `json:"oldState"`
	NewName          *string                                    `json:"newName,omitempty"`
	NewFlattenedName *BlockStateUpgradeSchemaModelFlattenInfo   `json:"newFlattenedName,omitempty"`
	NewState         map[string]BlockStateUpgradeSchemaModelTag `json:"newState"`
	CopiedState      []string                                   `json:"copiedState,omitempty"`
}

// BlockStateUpgradeSchemaModel is a port of BlockStateUpgradeSchemaModel.
type BlockStateUpgradeSchemaModel struct {
	MaxVersionMajor    *int `json:"maxVersionMajor"`
	MaxVersionMinor    *int `json:"maxVersionMinor"`
	MaxVersionPatch    *int `json:"maxVersionPatch"`
	MaxVersionRevision *int `json:"maxVersionRevision"`

	RenamedIds                  map[string]string                                     `json:"renamedIds,omitempty"`
	AddedProperties             map[string]map[string]BlockStateUpgradeSchemaModelTag `json:"addedProperties,omitempty"`
	RemovedProperties           map[string][]string                                   `json:"removedProperties,omitempty"`
	RenamedProperties           map[string]map[string]string                          `json:"renamedProperties,omitempty"`
	RemappedPropertyValues      map[string]map[string]string                          `json:"remappedPropertyValues,omitempty"`
	RemappedPropertyValuesIndex map[string][]BlockStateUpgradeSchemaModelValueRemap   `json:"remappedPropertyValuesIndex,omitempty"`
	FlattenedProperties         map[string]BlockStateUpgradeSchemaModelFlattenInfo    `json:"flattenedProperties,omitempty"`
	RemappedStates              map[string][]BlockStateUpgradeSchemaModelBlockRemap   `json:"remappedStates,omitempty"`
}
