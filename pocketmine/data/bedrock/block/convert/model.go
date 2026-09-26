package blockconvert

import "pocketmine-go/pocketmine/block"

// Model is a port of pocketmine\data\bedrock\block\convert\Model: a block with one Bedrock ID and
// a list of state properties.
type Model struct {
	block      block.Behavior
	id         string
	properties []Property
}

// NewModel is Model::create.
func NewModel(blk block.Behavior, id string) *Model { return &Model{block: blk, id: id} }

func (m *Model) GetBlock() block.Behavior  { return m.block }
func (m *Model) GetID() string             { return m.id }
func (m *Model) GetProperties() []Property { return m.properties }

// Properties is Model::properties.
func (m *Model) Properties(properties ...Property) *Model {
	m.properties = properties
	return m
}

// FlattenedIdModel is a port of pocketmine\data\bedrock\block\convert\FlattenedIdModel: a block
// whose Bedrock ID is built from several components (strings and StringProperty values), plus
// state properties.
type FlattenedIdModel struct {
	block        block.Behavior
	idComponents []any
	properties   []Property
}

// NewFlattenedIdModel is FlattenedIdModel::create.
func NewFlattenedIdModel(blk block.Behavior) *FlattenedIdModel { return &FlattenedIdModel{block: blk} }

func (m *FlattenedIdModel) GetBlock() block.Behavior  { return m.block }
func (m *FlattenedIdModel) GetIdComponents() []any    { return m.idComponents }
func (m *FlattenedIdModel) GetProperties() []Property { return m.properties }

// IdComponents is FlattenedIdModel::idComponents: each component is a string or a StringProperty.
func (m *FlattenedIdModel) IdComponents(components ...any) *FlattenedIdModel {
	for _, c := range components {
		switch c.(type) {
		case string, StringProperty:
		default:
			panic("ID components must be strings or StringProperty values")
		}
	}
	m.idComponents = components
	return m
}

// Properties is FlattenedIdModel::properties.
func (m *FlattenedIdModel) Properties(properties ...Property) *FlattenedIdModel {
	m.properties = properties
	return m
}

// concat is PHP's [...$a, $b, ...] for ID component lists.
func concat(parts ...any) []any {
	var result []any
	for _, part := range parts {
		if list, ok := part.([]any); ok {
			result = append(result, list...)
		} else {
			result = append(result, part)
		}
	}
	return result
}
