package itemupgrade

import (
	"fmt"
	"sort"
)

// ItemIdMetaUpgrader is a port of pocketmine\data\bedrock\item\upgrade\ItemIdMetaUpgrader.
type ItemIdMetaUpgrader struct {
	idMetaUpgradeSchemas map[int]*ItemIdMetaUpgradeSchema
	order                []int
}

// NewItemIdMetaUpgrader is a port of ItemIdMetaUpgrader::__construct.
func NewItemIdMetaUpgrader(schemas []*ItemIdMetaUpgradeSchema) *ItemIdMetaUpgrader {
	u := &ItemIdMetaUpgrader{idMetaUpgradeSchemas: map[int]*ItemIdMetaUpgradeSchema{}}
	for _, schema := range schemas {
		u.AddSchema(schema)
	}
	return u
}

// AddSchema is a port of ItemIdMetaUpgrader::addSchema. Panics if a schema with the same ID was
// already added.
func (u *ItemIdMetaUpgrader) AddSchema(schema *ItemIdMetaUpgradeSchema) {
	if _, ok := u.idMetaUpgradeSchemas[schema.GetSchemaID()]; ok {
		panic(fmt.Sprintf("Already have a schema with priority %d", schema.GetSchemaID()))
	}
	u.idMetaUpgradeSchemas[schema.GetSchemaID()] = schema
	u.order = append(u.order, schema.GetSchemaID())
	sort.Ints(u.order)
}

// GetSchemas is a port of ItemIdMetaUpgrader::getSchemas (ordered by schema ID).
func (u *ItemIdMetaUpgrader) GetSchemas() []*ItemIdMetaUpgradeSchema {
	result := make([]*ItemIdMetaUpgradeSchema, len(u.order))
	for i, id := range u.order {
		result[i] = u.idMetaUpgradeSchemas[id]
	}
	return result
}

// Upgrade is a port of ItemIdMetaUpgrader::upgrade.
func (u *ItemIdMetaUpgrader) Upgrade(id string, meta int) (string, int) {
	newID, newMeta := id, meta
	for _, schemaID := range u.order {
		schema := u.idMetaUpgradeSchemas[schemaID]
		if remapped, ok := schema.RemapMeta(newID, newMeta); ok {
			newID, newMeta = remapped, 0
		} else if renamed, ok := schema.RenameID(newID); ok {
			newID = renamed
		}
	}
	return newID, newMeta
}
