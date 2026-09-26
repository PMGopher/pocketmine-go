package bedrock

import "fmt"

// IntSaveIdMap is a port of pocketmine\data\bedrock\IntSaveIdMapTrait - a bidirectional map between
// saved integer IDs and identity-compared registry values (effects, enchantments, ...).
type IntSaveIdMap[T comparable] struct {
	idToEnum map[int]T
	enumToId map[T]int
}

func newIntSaveIdMap[T comparable]() *IntSaveIdMap[T] {
	return &IntSaveIdMap[T]{idToEnum: map[int]T{}, enumToId: map[T]int{}}
}

func (m *IntSaveIdMap[T]) Register(saveID int, enum T) {
	m.idToEnum[saveID] = enum
	m.enumToId[enum] = saveID
}

// FromID returns the value registered for id (we might not have all the IDs registered).
func (m *IntSaveIdMap[T]) FromID(id int) (T, bool) {
	v, ok := m.idToEnum[id]
	return v, ok
}

// ToID returns the save ID of enum. Panics if it has none, like PHP's InvalidArgumentException -
// every registry value is expected to be mapped.
func (m *IntSaveIdMap[T]) ToID(enum T) int {
	id, ok := m.enumToId[enum]
	if !ok {
		panic(fmt.Sprintf("bedrock: %v does not have a mapped save ID", enum))
	}
	return id
}

// NewIntSaveIdMap creates an empty IntSaveIdMap (for maps defined in other packages).
func NewIntSaveIdMap[T comparable]() *IntSaveIdMap[T] { return newIntSaveIdMap[T]() }
