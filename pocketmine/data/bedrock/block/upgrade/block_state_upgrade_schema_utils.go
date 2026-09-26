package blockupgrade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/data/bedrock/block/upgrade/model"
)

// Describe is a port of BlockStateUpgradeSchemaUtils::describe.
func Describe(schema *BlockStateUpgradeSchema) string {
	var lines []string
	lines = append(lines, "Renames:")
	for _, k := range sortedKeys(schema.RenamedIds) {
		lines = append(lines, "- "+schema.RenamedIds[k])
	}
	lines = append(lines, "Added properties:")
	for _, blockName := range sortedKeys(schema.AddedProperties) {
		tags := schema.AddedProperties[blockName]
		for _, k := range sortedKeys(tags) {
			lines = append(lines, fmt.Sprintf("- %s has %s added: %s", blockName, k, describeValue(tags[k])))
		}
	}
	lines = append(lines, "Removed properties:")
	for _, blockName := range sortedKeys(schema.RemovedProperties) {
		for _, tagName := range schema.RemovedProperties[blockName] {
			lines = append(lines, fmt.Sprintf("- %s has %s removed", blockName, tagName))
		}
	}
	lines = append(lines, "Renamed properties:")
	for _, blockName := range sortedKeys(schema.RenamedProperties) {
		names := schema.RenamedProperties[blockName]
		for _, oldName := range sortedKeys(names) {
			lines = append(lines, fmt.Sprintf("- %s has %s renamed to %s", blockName, oldName, names[oldName]))
		}
	}
	lines = append(lines, "Remapped property values:")
	for _, blockName := range sortedKeys(schema.RemappedPropertyValues) {
		remaps := schema.RemappedPropertyValues[blockName]
		for _, tagName := range sortedKeys(remaps) {
			for _, oldNew := range remaps[tagName] {
				lines = append(lines, fmt.Sprintf("- %s has %s value changed from %s to %s", blockName, tagName, describeValue(oldNew.Old), describeValue(oldNew.New)))
			}
		}
	}
	return strings.Join(lines, "\n")
}

// describeValue is Tag::__toString for a state value.
func describeValue(v any) string {
	switch v := v.(type) {
	case int32:
		return fmt.Sprintf("TAG_Int=%d", v)
	case uint8:
		return fmt.Sprintf("TAG_Byte=%d", int8(v))
	case string:
		return fmt.Sprintf("TAG_String=\"%s\"", v)
	}
	return fmt.Sprint(v)
}

// jsonModelToTag is a port of BlockStateUpgradeSchemaUtils::jsonModelToTag.
func jsonModelToTag(m model.BlockStateUpgradeSchemaModelTag) (any, error) {
	switch {
	case m.Byte != nil && m.Int == nil && m.String == nil:
		return uint8(int8(*m.Byte)), nil
	case m.Byte == nil && m.Int != nil && m.String == nil:
		return int32(*m.Int), nil
	case m.Byte == nil && m.Int == nil && m.String != nil:
		return *m.String, nil
	}
	return nil, fmt.Errorf("Malformed JSON model tag, expected exactly one of 'byte', 'int' or 'string' properties")
}

func jsonModelToTags(m map[string]model.BlockStateUpgradeSchemaModelTag) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		t, err := jsonModelToTag(v)
		if err != nil {
			return nil, err
		}
		result[k] = t
	}
	return result, nil
}

// FromJsonModel is a port of BlockStateUpgradeSchemaUtils::fromJsonModel.
func FromJsonModel(m *model.BlockStateUpgradeSchemaModel, schemaID int) (*BlockStateUpgradeSchema, error) {
	if m.MaxVersionMajor == nil || m.MaxVersionMinor == nil || m.MaxVersionPatch == nil || m.MaxVersionRevision == nil {
		return nil, fmt.Errorf("missing required maxVersion* properties")
	}
	result := NewBlockStateUpgradeSchema(*m.MaxVersionMajor, *m.MaxVersionMinor, *m.MaxVersionPatch, *m.MaxVersionRevision, schemaID)
	if m.RenamedIds != nil {
		result.RenamedIds = m.RenamedIds
	}
	if m.RenamedProperties != nil {
		result.RenamedProperties = m.RenamedProperties
	}
	if m.RemovedProperties != nil {
		result.RemovedProperties = m.RemovedProperties
	}

	for blockName, properties := range m.AddedProperties {
		tags, err := jsonModelToTags(properties)
		if err != nil {
			return nil, err
		}
		result.AddedProperties[blockName] = tags
	}

	convertedRemappedValuesIndex := map[string][]BlockStateUpgradeSchemaValueRemap{}
	for mappingKey, mappingValues := range m.RemappedPropertyValuesIndex {
		for _, oldNew := range mappingValues {
			oldTag, err := jsonModelToTag(oldNew.Old)
			if err != nil {
				return nil, err
			}
			newTag, err := jsonModelToTag(oldNew.New)
			if err != nil {
				return nil, err
			}
			convertedRemappedValuesIndex[mappingKey] = append(convertedRemappedValuesIndex[mappingKey], BlockStateUpgradeSchemaValueRemap{Old: oldTag, New: newTag})
		}
	}

	for blockName, properties := range m.RemappedPropertyValues {
		for property, mappedValuesKey := range properties {
			values, ok := convertedRemappedValuesIndex[mappedValuesKey]
			if !ok {
				return nil, fmt.Errorf("Missing key from schema values index %s", mappedValuesKey)
			}
			if result.RemappedPropertyValues[blockName] == nil {
				result.RemappedPropertyValues[blockName] = map[string][]BlockStateUpgradeSchemaValueRemap{}
			}
			result.RemappedPropertyValues[blockName][property] = values
		}
	}

	for blockName, flattenRule := range m.FlattenedProperties {
		rule, err := jsonModelToFlattenRule(flattenRule)
		if err != nil {
			return nil, err
		}
		result.FlattenedProperties[blockName] = rule
	}

	for oldBlockName, remaps := range m.RemappedStates {
		for _, remap := range remaps {
			converted := BlockStateUpgradeSchemaBlockRemap{CopiedState: remap.CopiedState}
			switch {
			case remap.NewName != nil:
				converted.NewName = *remap.NewName
			case remap.NewFlattenedName != nil:
				rule, err := jsonModelToFlattenRule(*remap.NewFlattenedName)
				if err != nil {
					return nil, err
				}
				converted.NewFlattenedName = rule
			default:
				return nil, fmt.Errorf("Expected exactly one of 'newName' or 'newFlattenedName' properties to be set")
			}
			var err error
			if converted.OldState, err = jsonModelToTags(remap.OldState); err != nil {
				return nil, err
			}
			if converted.NewState, err = jsonModelToTags(remap.NewState); err != nil {
				return nil, err
			}
			result.RemappedStates[oldBlockName] = append(result.RemappedStates[oldBlockName], converted)
		}
	}

	return result, nil
}

// jsonModelToFlattenRule is a port of BlockStateUpgradeSchemaUtils::jsonModelToFlattenRule.
func jsonModelToFlattenRule(m model.BlockStateUpgradeSchemaModelFlattenInfo) (*BlockStateUpgradeSchemaFlattenInfo, error) {
	typ := FlattenedTypeString
	if m.FlattenedPropertyType != nil {
		switch *m.FlattenedPropertyType {
		case "string":
		case "int":
			typ = FlattenedTypeInt
		case "byte":
			typ = FlattenedTypeByte
		default:
			return nil, fmt.Errorf("Unexpected flattened property type %s, expected 'string', 'int' or 'byte'", *m.FlattenedPropertyType)
		}
	}
	remaps := m.FlattenedValueRemaps
	if remaps == nil {
		remaps = map[string]string{}
	}
	return &BlockStateUpgradeSchemaFlattenInfo{
		Prefix:                m.Prefix,
		FlattenedProperty:     m.FlattenedProperty,
		Suffix:                m.Suffix,
		FlattenedValueRemaps:  remaps,
		FlattenedPropertyType: typ,
	}, nil
}

var schemaFileRegex = regexp.MustCompile(`^(\d{4}).*\.json$`)

// LoadSchemas is a port of BlockStateUpgradeSchemaUtils::loadSchemas: every schema file in dir
// (of fsys) with a schema ID up to maxSchemaID, ordered by schema ID (oldest first).
func LoadSchemas(fsys fs.FS, dir string, maxSchemaID int) ([]*BlockStateUpgradeSchema, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var result []*BlockStateUpgradeSchema
	for _, entry := range entries {
		matches := schemaFileRegex.FindStringSubmatch(entry.Name())
		if matches == nil || entry.IsDir() {
			continue
		}
		schemaID, _ := strconv.Atoi(matches[1])
		if schemaID > maxSchemaID {
			continue
		}
		fullPath := path.Join(dir, entry.Name())
		raw, err := fs.ReadFile(fsys, fullPath)
		if err != nil {
			return nil, err
		}
		schema, err := LoadSchemaFromString(raw, schemaID)
		if err != nil {
			return nil, fmt.Errorf("Loading schema file %s: %w", fullPath, err)
		}
		result = append(result, schema)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GetSchemaID() < result[j].GetSchemaID() })
	return result, nil
}

// LoadSchemaFromString is a port of BlockStateUpgradeSchemaUtils::loadSchemaFromString. Unknown
// properties are an error, like PHP's JsonMapper with bExceptionOnUndefinedProperty.
func LoadSchemaFromString(raw []byte, schemaID int) (*BlockStateUpgradeSchema, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var m model.BlockStateUpgradeSchemaModel
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	return FromJsonModel(&m, schemaID)
}
