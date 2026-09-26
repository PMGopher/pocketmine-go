package itemupgrade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"

	"pocketmine-go/pocketmine/data/bedrock/item/upgrade/model"
)

var schemaFileRegex = regexp.MustCompile(`^(\d{4}).*\.json$`)

// LoadSchemas is a port of ItemIdMetaUpgradeSchemaUtils::loadSchemas: every schema file in dir (of
// fsys) with a schema ID up to maxSchemaID, ordered by schema ID.
func LoadSchemas(fsys fs.FS, dir string, maxSchemaID int) ([]*ItemIdMetaUpgradeSchema, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var result []*ItemIdMetaUpgradeSchema
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

// LoadSchemaFromString is a port of ItemIdMetaUpgradeSchemaUtils::loadSchemaFromString.
func LoadSchemaFromString(raw []byte, schemaID int) (*ItemIdMetaUpgradeSchema, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var m model.ItemIdMetaUpgradeSchemaModel
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	renamed := m.RenamedIds
	if renamed == nil {
		renamed = map[string]string{}
	}
	remapped := make(map[string]map[int]string, len(m.RemappedMetas))
	for id, metas := range m.RemappedMetas {
		remapped[id] = make(map[int]string, len(metas))
		for metaKey, newID := range metas {
			meta, err := strconv.Atoi(metaKey)
			if err != nil {
				return nil, fmt.Errorf("invalid meta key %q for %s", metaKey, id)
			}
			remapped[id][meta] = newID
		}
	}
	return NewItemIdMetaUpgradeSchema(renamed, remapped, schemaID), nil
}
