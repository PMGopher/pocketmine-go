package nbt

// Type is a port of the NBT::TAG_* constants.
type Type uint8

const (
	TagEnd       Type = 0
	TagByte      Type = 1
	TagShort     Type = 2
	TagInt       Type = 3
	TagLong      Type = 4
	TagFloat     Type = 5
	TagDouble    Type = 6
	TagByteArray Type = 7
	TagString    Type = 8
	TagList      Type = 9
	TagCompound  Type = 10
	TagIntArray  Type = 11
)

// Tag is a port of pocketmine\nbt\tag\Tag.
//
// PHP's scalar tags (Byte/Short/Int/Long/Float/Double/String/ByteArray/IntArray) are all
// ImmutableTag subclasses whose only job is to hold a value and range-check it once at
// construction. Here they're plain Go value types (ByteTag int8, and so on) instead of structs:
// Go's int8/int16/int32/int64 already enforce the byte/short/int/long ranges at the type level
// (an out-of-range literal is a compile error, an out-of-range runtime value can't exist), so the
// PHP original's min()/max() checks and the whole safeClone()/makeCopy()/cloning-guard machinery
// (needed because PHP objects are reference types) both disappear — a Go value type is already
// its own independent copy on every assignment.
//
// Only CompoundTag and ListTag (which hold mutable slices/maps of child tags) are structs with
// real Clone() methods, mirroring PHP's __clone()/safeClone() recursion for container tags.
type Tag interface {
	Type() Type
	Equals(other Tag) bool
	String() string

	write(w StreamWriter) error
	stringify(indentation int) string
}

func typeName(t Type) string {
	switch t {
	case TagEnd:
		return "End"
	case TagByte:
		return "Byte"
	case TagShort:
		return "Short"
	case TagInt:
		return "Int"
	case TagLong:
		return "Long"
	case TagFloat:
		return "Float"
	case TagDouble:
		return "Double"
	case TagByteArray:
		return "ByteArray"
	case TagString:
		return "String"
	case TagList:
		return "List"
	case TagCompound:
		return "Compound"
	case TagIntArray:
		return "IntArray"
	default:
		return "Unknown"
	}
}
