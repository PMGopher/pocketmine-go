package blockconvert

import "fmt"

// rawValue is the type of a raw block state property value: int for IntTag properties, string for
// StringTag properties.
type rawValue interface{ ~int | ~string }

// pair is one entry of a PHP [value => raw] map literal.
type pair[V comparable, R rawValue] struct {
	value V
	raw   R
}

func p[V comparable, R rawValue](value V, raw R) pair[V, R] { return pair[V, R]{value, raw} }

// ValueMap is a port of IntFromRawStateMap and EnumFromRawStateMap (property\StateMap): a two-way
// map between a Go value (an int, a facing, an enum case) and its raw state value. Deserialize
// aliases map extra raw values onto a value.
type ValueMap[V comparable, R rawValue] struct {
	serialize   map[V]R
	deserialize map[R]V
	// rawOrder is the deserialize map's keys in PHP's insertion order (getRawToValueMap()).
	rawOrder []R
}

// newValueMap is IntFromRawStateMap/EnumFromRawStateMap::__construct, entries in PHP's order.
func newValueMap[V comparable, R rawValue](entries ...pair[V, R]) *ValueMap[V, R] {
	m := &ValueMap[V, R]{serialize: map[V]R{}, deserialize: map[R]V{}}
	for _, e := range entries {
		m.serialize[e.value] = e.raw
		m.addRaw(e.raw, e.value)
	}
	return m
}

func (m *ValueMap[V, R]) addRaw(raw R, value V) {
	if _, exists := m.deserialize[raw]; !exists {
		m.rawOrder = append(m.rawOrder, raw)
	}
	m.deserialize[raw] = value
}

// withAliases adds deserialize aliases (IntFromRawStateMap's $deserializeAliases / EnumFromRawStateMap's alias mapper).
func (m *ValueMap[V, R]) withAliases(aliases ...pair[V, R]) *ValueMap[V, R] {
	for _, a := range aliases {
		m.addRaw(a.raw, a.value)
	}
	return m
}

// ValueToRaw is StateMap::valueToRaw.
func (m *ValueMap[V, R]) ValueToRaw(value V) R {
	raw, ok := m.serialize[value]
	if !ok {
		serializeError("No raw state value mapped for %v", value)
	}
	return raw
}

// RawToValue is StateMap::rawToValue.
func (m *ValueMap[V, R]) RawToValue(raw R) (V, bool) {
	value, ok := m.deserialize[raw]
	return value, ok
}

// RawValues is array_keys(getRawToValueMap()).
func (m *ValueMap[V, R]) RawValues() []R { return append([]R(nil), m.rawOrder...) }

func (m *ValueMap[V, R]) printableValue(value V) string { return fmt.Sprint(value) }
