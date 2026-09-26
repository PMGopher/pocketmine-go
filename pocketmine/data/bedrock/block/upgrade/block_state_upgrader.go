package blockupgrade

import (
	"fmt"
	"sort"
	"strconv"

	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockStateUpgrader is a port of pocketmine\data\bedrock\block\upgrade\BlockStateUpgrader.
type BlockStateUpgrader struct {
	// upgradeSchemas is versionId => [schemaId => schema] (PHP keeps both levels ksorted; the
	// sorted keys are kept alongside).
	upgradeSchemas map[int32]map[int]*BlockStateUpgradeSchema
	versionOrder   []int32
	schemaOrder    map[int32][]int
	outputVersion  int32
}

// NewBlockStateUpgrader is a port of BlockStateUpgrader::__construct.
func NewBlockStateUpgrader(schemas []*BlockStateUpgradeSchema) *BlockStateUpgrader {
	u := &BlockStateUpgrader{
		upgradeSchemas: map[int32]map[int]*BlockStateUpgradeSchema{},
		schemaOrder:    map[int32][]int{},
	}
	for _, schema := range schemas {
		u.AddSchema(schema)
	}
	return u
}

// AddSchema is a port of BlockStateUpgrader::addSchema. Panics if a schema with the same schema
// ID and version ID was already added (PHP's InvalidArgumentException).
func (u *BlockStateUpgrader) AddSchema(schema *BlockStateUpgradeSchema) {
	schemaID, versionID := schema.GetSchemaID(), schema.GetVersionID()
	if _, ok := u.upgradeSchemas[versionID][schemaID]; ok {
		panic("Cannot add two schemas with the same schema ID and version ID")
	}
	if u.upgradeSchemas[versionID] == nil {
		u.upgradeSchemas[versionID] = map[int]*BlockStateUpgradeSchema{}
		u.versionOrder = append(u.versionOrder, versionID)
		sort.Slice(u.versionOrder, func(i, j int) bool { return u.versionOrder[i] < u.versionOrder[j] })
	}
	// The schema ID tells us the order when multiple schemas use the same version ID.
	u.upgradeSchemas[versionID][schemaID] = schema
	u.schemaOrder[versionID] = append(u.schemaOrder[versionID], schemaID)
	sort.Ints(u.schemaOrder[versionID])

	u.outputVersion = max(u.outputVersion, versionID)
}

// Upgrade is a port of BlockStateUpgrader::upgrade.
func (u *BlockStateUpgrader) Upgrade(data bedrock.BlockStateData) bedrock.BlockStateData {
	version := data.Version
	name := data.Name
	states := make(map[string]any, len(data.States))
	for k, v := range data.States {
		states[k] = v
	}
	for _, resultVersion := range u.versionOrder {
		schemaList := u.schemaOrder[resultVersion]
		// Sometimes Mojang made changes without bumping the version ID (e.g.
		// 0131_1.18.20.27_beta_to_1.18.30.json renamed a bunch of block IDs). When that happens, all
		// schemas of that version must be applied even if the version is the same, since the input
		// version doesn't tell us which of them were already applied. If there's only one schema for
		// a version (the norm), it's safe to assume it was already applied.
		if version > resultVersion || (len(schemaList) == 1 && version == resultVersion) {
			continue
		}
		for _, schemaID := range schemaList {
			name, states = u.applySchema(u.upgradeSchemas[resultVersion][schemaID], name, states)
		}
	}
	return bedrock.BlockStateData{Name: name, States: states, Version: u.outputVersion}
}

// GetOutputVersion is the version of the blockstates Upgrade returns.
func (u *BlockStateUpgrader) GetOutputVersion() int32 { return u.outputVersion }

func (u *BlockStateUpgrader) applySchema(schema *BlockStateUpgradeSchema, oldName string, states map[string]any) (string, map[string]any) {
	if newName, newStates, ok := u.applyStateRemapped(schema, oldName, states); ok {
		return newName, newStates
	}

	renamed, hasRename := schema.RenamedIds[oldName]
	flatten, hasFlatten := schema.FlattenedProperties[oldName]
	if hasRename && hasFlatten {
		panic(fmt.Sprintf("Both renamedIds and flattenedProperties are set for the same block ID \"%s\" - don't know what to do", oldName))
	}
	newName := oldName
	if hasRename {
		newName = renamed
	} else if hasFlatten {
		newName, states = u.applyPropertyFlattened(flatten, oldName, states)
	}

	states = u.applyPropertyAdded(schema, oldName, states)
	states = u.applyPropertyRemoved(schema, oldName, states)
	states = u.applyPropertyRenamedOrValueChanged(schema, oldName, states)
	states = u.applyPropertyValueChanged(schema, oldName, states)

	return newName, states
}

func (u *BlockStateUpgrader) applyStateRemapped(schema *BlockStateUpgradeSchema, oldName string, oldState map[string]any) (string, map[string]any, bool) {
	remaps, ok := schema.RemappedStates[oldName]
	if !ok {
		return "", nil, false
	}
nextRemap:
	for _, remap := range remaps {
		if len(remap.OldState) > len(oldState) {
			// match criteria has more requirements than we have state properties
			continue
		}
		for k, v := range remap.OldState {
			if have, ok := oldState[k]; !ok || !bedrock.StateValuesEqual(have, v) {
				continue nextRemap
			}
		}

		newName := remap.NewName
		if remap.NewFlattenedName != nil {
			// discard flatten modifications to state - the remap newState and copiedState take care of it
			newName, _ = u.applyPropertyFlattened(remap.NewFlattenedName, oldName, oldState)
		}

		newState := make(map[string]any, len(remap.NewState)+len(remap.CopiedState))
		for k, v := range remap.NewState {
			newState[k] = v
		}
		for _, stateName := range remap.CopiedState {
			if v, ok := oldState[stateName]; ok {
				newState[stateName] = v
			}
		}
		return newName, newState, true
	}
	return "", nil, false
}

func (u *BlockStateUpgrader) applyPropertyAdded(schema *BlockStateUpgradeSchema, oldName string, states map[string]any) map[string]any {
	for propertyName, value := range schema.AddedProperties[oldName] {
		if _, ok := states[propertyName]; !ok {
			states[propertyName] = value
		}
	}
	return states
}

func (u *BlockStateUpgrader) applyPropertyRemoved(schema *BlockStateUpgradeSchema, oldName string, states map[string]any) map[string]any {
	for _, propertyName := range schema.RemovedProperties[oldName] {
		delete(states, propertyName)
	}
	return states
}

func (u *BlockStateUpgrader) locateNewPropertyValue(schema *BlockStateUpgradeSchema, oldName, oldPropertyName string, oldValue any) any {
	for _, pair := range schema.RemappedPropertyValues[oldName][oldPropertyName] {
		if bedrock.StateValuesEqual(pair.Old, oldValue) {
			return pair.New
		}
	}
	return oldValue
}

func (u *BlockStateUpgrader) applyPropertyRenamedOrValueChanged(schema *BlockStateUpgradeSchema, oldName string, states map[string]any) map[string]any {
	renames, ok := schema.RenamedProperties[oldName]
	if !ok {
		return states
	}
	// PHP iterates the renames in schema (insertion) order; renames never chain within one schema,
	// so name order gives the same result.
	for _, oldPropertyName := range sortedKeys(renames) {
		oldValue, ok := states[oldPropertyName]
		if !ok {
			continue
		}
		delete(states, oldPropertyName)
		// A value remap has to be done here, since it's indexed by the old property name.
		states[renames[oldPropertyName]] = u.locateNewPropertyValue(schema, oldName, oldPropertyName, oldValue)
	}
	return states
}

func (u *BlockStateUpgrader) applyPropertyValueChanged(schema *BlockStateUpgradeSchema, oldName string, states map[string]any) map[string]any {
	for oldPropertyName := range schema.RemappedPropertyValues[oldName] {
		if oldValue, ok := states[oldPropertyName]; ok {
			states[oldPropertyName] = u.locateNewPropertyValue(schema, oldName, oldPropertyName, oldValue)
		}
	}
	return states
}

func (u *BlockStateUpgrader) applyPropertyFlattened(flattenInfo *BlockStateUpgradeSchemaFlattenInfo, oldName string, states map[string]any) (string, map[string]any) {
	flattenedValue, ok := states[flattenInfo.FlattenedProperty]
	if !ok || !flattenInfo.FlattenedPropertyType.matches(flattenedValue) {
		// flattened property is not of the expected type, so this transformation is not applicable
		return oldName, states
	}
	var embedKey string
	switch v := flattenedValue.(type) {
	case string:
		embedKey = v
	case uint8:
		embedKey = strconv.Itoa(int(int8(v)))
	case int32:
		embedKey = strconv.Itoa(int(v))
	}
	embedValue := embedKey
	if remapped, ok := flattenInfo.FlattenedValueRemaps[embedKey]; ok {
		embedValue = remapped
	}
	newName := flattenInfo.Prefix + embedValue + flattenInfo.Suffix
	delete(states, flattenInfo.FlattenedProperty)
	return newName, states
}
