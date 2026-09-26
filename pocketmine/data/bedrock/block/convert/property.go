package blockconvert

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/block"
)

// Property is a port of pocketmine\data\bedrock\block\convert\property\Property: one block state
// property, read from / written to a block through getter and setter closures.
type Property interface {
	GetName() string
	Deserialize(blk block.Behavior, in *BlockStateReader)
	Serialize(blk block.Behavior, out *BlockStateWriter)
}

// StringProperty is a port of property\StringProperty: a string property that can also be used as
// part of a flattened block ID.
type StringProperty interface {
	Property
	GetPossibleValues() []string
	DeserializePlain(blk block.Behavior, raw string)
	SerializePlain(blk block.Behavior) string
}

// as asserts blk to the closure's block type (PHP's closure parameter type).
func as[B any](blk block.Behavior) B {
	b, ok := blk.(B)
	if !ok {
		var zero B
		panic(fmt.Sprintf("blockconvert: %T is not a %T", blk, zero))
	}
	return b
}

// BoolProperty is a port of property\BoolProperty.
type BoolProperty[B any] struct {
	name     string
	getter   func(B) bool
	setter   func(B, bool)
	inverted bool //we don't *need* this, but it avoids accidentally forgetting a ! in the getter/setter closures (and makes it analysable)
}

func NewBoolProperty[B any](name string, getter func(B) bool, setter func(B, bool)) *BoolProperty[B] {
	return &BoolProperty[B]{name: name, getter: getter, setter: setter}
}

// NewInvertedBoolProperty is `new BoolProperty(..., inverted: true)`.
func NewInvertedBoolProperty[B any](name string, getter func(B) bool, setter func(B, bool)) *BoolProperty[B] {
	return &BoolProperty[B]{name: name, getter: getter, setter: setter, inverted: true}
}

// UnusedBoolProperty is a port of BoolProperty::unused.
func UnusedBoolProperty(name string, serializedValue bool) *BoolProperty[block.Behavior] {
	return NewBoolProperty(name, func(block.Behavior) bool { return serializedValue }, nil)
}

func (p *BoolProperty[B]) GetName() string { return p.name }

func (p *BoolProperty[B]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	raw := in.ReadBool(p.name)
	if p.setter != nil {
		p.setter(as[B](blk), raw != p.inverted)
	}
}

func (p *BoolProperty[B]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	value := p.getter(as[B](blk))
	out.WriteBool(p.name, value != p.inverted)
}

// IntProperty is a port of property\IntProperty.
type IntProperty[B any] struct {
	name     string
	min, max int
	getter   func(B) int
	setter   func(B, int)
	offset   int
}

func NewIntProperty[B any](name string, min, max int, getter func(B) int, setter func(B, int)) *IntProperty[B] {
	return NewIntPropertyWithOffset(name, min, max, getter, setter, 0)
}

// NewIntPropertyWithOffset is `new IntProperty(..., offset: $offset)`.
func NewIntPropertyWithOffset[B any](name string, min, max int, getter func(B) int, setter func(B, int), offset int) *IntProperty[B] {
	if min > max {
		panic("Min value cannot be greater than max value")
	}
	return &IntProperty[B]{name: name, min: min, max: max, getter: getter, setter: setter, offset: offset}
}

// UnusedIntProperty is a port of IntProperty::unused.
func UnusedIntProperty(name string, serializedValue int) *IntProperty[block.Behavior] {
	return NewIntProperty(name, stdmath.MinInt32, stdmath.MaxInt32, func(block.Behavior) int { return serializedValue }, nil)
}

func (p *IntProperty[B]) GetName() string { return p.name }

func (p *IntProperty[B]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	value := in.ReadBoundedInt(p.name, p.min, p.max)
	if p.setter != nil {
		p.setter(as[B](blk), value+p.offset)
	}
}

func (p *IntProperty[B]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	value := p.getter(as[B](blk))
	out.WriteInt(p.name, value-p.offset)
}

// DummyProperty is a port of property\DummyProperty: a property with a fixed value that isn't
// stored on the block.
type DummyProperty struct {
	name  string
	value any // bool, int or string
}

func NewDummyProperty(name string, value any) *DummyProperty {
	switch value.(type) {
	case bool, int, string:
	default:
		panic("DummyProperty value must be bool, int or string")
	}
	return &DummyProperty{name: name, value: value}
}

func (p *DummyProperty) GetName() string { return p.name }

func (p *DummyProperty) Deserialize(blk block.Behavior, in *BlockStateReader) { in.Ignored(p.name) }

func (p *DummyProperty) Serialize(blk block.Behavior, out *BlockStateWriter) {
	switch v := p.value.(type) {
	case bool:
		out.WriteBool(p.name, v)
	case int:
		out.WriteInt(p.name, v)
	case string:
		out.WriteString(p.name, v)
	}
}

// BoolFromStringProperty is a port of property\BoolFromStringProperty.
type BoolFromStringProperty[B any] struct {
	name       string
	falseValue string
	trueValue  string
	getter     func(B) bool
	setter     func(B, bool)
}

func NewBoolFromStringProperty[B any](name, falseValue, trueValue string, getter func(B) bool, setter func(B, bool)) *BoolFromStringProperty[B] {
	return &BoolFromStringProperty[B]{name: name, falseValue: falseValue, trueValue: trueValue, getter: getter, setter: setter}
}

func (p *BoolFromStringProperty[B]) GetName() string { return p.name }

func (p *BoolFromStringProperty[B]) GetPossibleValues() []string {
	return []string{p.falseValue, p.trueValue}
}

func (p *BoolFromStringProperty[B]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	p.DeserializePlain(blk, in.ReadString(p.name))
}

func (p *BoolFromStringProperty[B]) DeserializePlain(blk block.Behavior, raw string) {
	var value bool
	switch raw {
	case p.falseValue:
		value = false
	case p.trueValue:
		value = true
	default:
		serializeError("Invalid value for %s: %s", p.name, raw)
	}
	if p.setter != nil {
		p.setter(as[B](blk), value)
	}
}

func (p *BoolFromStringProperty[B]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	out.WriteString(p.name, p.SerializePlain(blk))
}

func (p *BoolFromStringProperty[B]) SerializePlain(blk block.Behavior) string {
	if p.getter(as[B](blk)) {
		return p.trueValue
	}
	return p.falseValue
}

// ValueFromStringProperty is a port of property\ValueFromStringProperty.
type ValueFromStringProperty[B any, V comparable] struct {
	name   string
	m      *ValueMap[V, string]
	getter func(B) V
	setter func(B, V)
}

func NewValueFromStringProperty[B any, V comparable](name string, m *ValueMap[V, string], getter func(B) V, setter func(B, V)) *ValueFromStringProperty[B, V] {
	return &ValueFromStringProperty[B, V]{name: name, m: m, getter: getter, setter: setter}
}

func (p *ValueFromStringProperty[B, V]) GetName() string { return p.name }

func (p *ValueFromStringProperty[B, V]) GetPossibleValues() []string { return p.m.RawValues() }

func (p *ValueFromStringProperty[B, V]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	p.DeserializePlain(blk, in.ReadString(p.name))
}

func (p *ValueFromStringProperty[B, V]) DeserializePlain(blk block.Behavior, raw string) {
	//TODO: duplicated code from BlockStateReader :(
	value, ok := p.m.RawToValue(raw)
	if !ok {
		panic(deserializeError("Property \"%s\" has invalid value \"%s\"", p.name, raw))
	}
	if p.setter != nil {
		p.setter(as[B](blk), value)
	}
}

func (p *ValueFromStringProperty[B, V]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	out.WriteString(p.name, p.SerializePlain(blk))
}

func (p *ValueFromStringProperty[B, V]) SerializePlain(blk block.Behavior) string {
	return p.m.ValueToRaw(p.getter(as[B](blk)))
}

// ValueFromIntProperty is a port of property\ValueFromIntProperty.
type ValueFromIntProperty[B any, V comparable] struct {
	name   string
	m      *ValueMap[V, int]
	getter func(B) V
	setter func(B, V)
}

func NewValueFromIntProperty[B any, V comparable](name string, m *ValueMap[V, int], getter func(B) V, setter func(B, V)) *ValueFromIntProperty[B, V] {
	return &ValueFromIntProperty[B, V]{name: name, m: m, getter: getter, setter: setter}
}

func (p *ValueFromIntProperty[B, V]) GetName() string { return p.name }

func (p *ValueFromIntProperty[B, V]) GetPossibleValues() []int { return p.m.RawValues() }

func (p *ValueFromIntProperty[B, V]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	raw := in.ReadInt(p.name)
	value, ok := p.m.RawToValue(raw)
	if !ok {
		panic(in.BadValueError(p.name, fmt.Sprint(raw), ""))
	}
	if p.setter != nil {
		p.setter(as[B](blk), value)
	}
}

func (p *ValueFromIntProperty[B, V]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	out.WriteInt(p.name, p.m.ValueToRaw(p.getter(as[B](blk))))
}

// ValueSetFromIntProperty is a port of property\ValueSetFromIntProperty: a set of values stored as
// bit flags.
type ValueSetFromIntProperty[B any, V comparable] struct {
	name     string
	m        *ValueMap[V, int]
	getter   func(B) []V
	setter   func(B, []V)
	maxValue int
}

func NewValueSetFromIntProperty[B any, V comparable](name string, m *ValueMap[V, int], getter func(B) []V, setter func(B, []V)) *ValueSetFromIntProperty[B, V] {
	p := &ValueSetFromIntProperty[B, V]{name: name, m: m, getter: getter, setter: setter}
	for _, possibleFlag := range m.RawValues() {
		option, _ := m.RawToValue(possibleFlag)
		if p.maxValue&possibleFlag != 0 {
			for _, otherFlag := range m.RawValues() {
				otherOption, _ := m.RawToValue(otherFlag)
				if possibleFlag&otherFlag == otherFlag && otherOption != option {
					panic(fmt.Sprintf("Flag for option %s overlaps with flag for option %s in property %s", m.printableValue(option), m.printableValue(otherOption), name))
				}
			}
			panic("Unreachable")
		}
		p.maxValue |= possibleFlag
	}
	return p
}

func (p *ValueSetFromIntProperty[B, V]) GetName() string { return p.name }

func (p *ValueSetFromIntProperty[B, V]) Deserialize(blk block.Behavior, in *BlockStateReader) {
	flags := in.ReadBoundedInt(p.name, 0, p.maxValue)
	var value []V
	for _, possibleFlag := range p.m.RawValues() {
		if flags&possibleFlag == possibleFlag {
			option, _ := p.m.RawToValue(possibleFlag)
			value = append(value, option)
		}
	}
	p.setter(as[B](blk), value)
}

func (p *ValueSetFromIntProperty[B, V]) Serialize(blk block.Behavior, out *BlockStateWriter) {
	flags := 0
	for _, option := range p.getter(as[B](blk)) {
		flags |= p.m.ValueToRaw(option)
	}
	out.WriteInt(p.name, flags)
}
