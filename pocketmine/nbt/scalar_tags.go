package nbt

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

func base64Encode(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

// ByteTag is a port of pocketmine\nbt\tag\ByteTag. See tag.go's doc comment for why this is a
// plain value type rather than a class with range-checking machinery.
type ByteTag int8

func (t ByteTag) Type() Type { return TagByte }
func (t ByteTag) Equals(other Tag) bool {
	o, ok := other.(ByteTag)
	return ok && o == t
}
func (t ByteTag) String() string             { return t.stringify(0) }
func (t ByteTag) stringify(int) string       { return "TAG_Byte=" + strconv.Itoa(int(t)) }
func (t ByteTag) write(w StreamWriter) error { w.WriteByte(uint8(t)); return nil }

type ShortTag int16

func (t ShortTag) Type() Type                 { return TagShort }
func (t ShortTag) Equals(other Tag) bool      { o, ok := other.(ShortTag); return ok && o == t }
func (t ShortTag) String() string             { return t.stringify(0) }
func (t ShortTag) stringify(int) string       { return "TAG_Short=" + strconv.Itoa(int(t)) }
func (t ShortTag) write(w StreamWriter) error { w.WriteShort(uint16(t)); return nil }

type IntTag int32

func (t IntTag) Type() Type                 { return TagInt }
func (t IntTag) Equals(other Tag) bool      { o, ok := other.(IntTag); return ok && o == t }
func (t IntTag) String() string             { return t.stringify(0) }
func (t IntTag) stringify(int) string       { return "TAG_Int=" + strconv.Itoa(int(t)) }
func (t IntTag) write(w StreamWriter) error { w.WriteInt(int32(t)); return nil }

type LongTag int64

func (t LongTag) Type() Type                 { return TagLong }
func (t LongTag) Equals(other Tag) bool      { o, ok := other.(LongTag); return ok && o == t }
func (t LongTag) String() string             { return t.stringify(0) }
func (t LongTag) stringify(int) string       { return "TAG_Long=" + strconv.FormatInt(int64(t), 10) }
func (t LongTag) write(w StreamWriter) error { w.WriteLong(int64(t)); return nil }

// FloatTag is a port of pocketmine\nbt\tag\FloatTag.
//
// PHP stores the value in a 64-bit float and has to special-case Equals() to compare via
// Binary::writeFloat() so extra 64-bit precision doesn't break comparisons after a round trip
// through the (32-bit) wire format. A genuine Go float32 has no such extra precision to begin
// with, so plain == already does the right thing.
type FloatTag float32

func (t FloatTag) Type() Type            { return TagFloat }
func (t FloatTag) Equals(other Tag) bool { o, ok := other.(FloatTag); return ok && o == t }
func (t FloatTag) String() string        { return t.stringify(0) }
func (t FloatTag) stringify(int) string {
	return "TAG_Float=" + strconv.FormatFloat(float64(t), 'g', -1, 32)
}
func (t FloatTag) write(w StreamWriter) error { w.WriteFloat(float32(t)); return nil }

type DoubleTag float64

func (t DoubleTag) Type() Type            { return TagDouble }
func (t DoubleTag) Equals(other Tag) bool { o, ok := other.(DoubleTag); return ok && o == t }
func (t DoubleTag) String() string        { return t.stringify(0) }
func (t DoubleTag) stringify(int) string {
	return "TAG_Double=" + strconv.FormatFloat(float64(t), 'g', -1, 64)
}
func (t DoubleTag) write(w StreamWriter) error { w.WriteDouble(float64(t)); return nil }

// StringTag is a port of pocketmine\nbt\tag\StringTag.
//
// Unlike the numeric tags, Go's string type has no built-in length cap, so the 32767-byte limit
// (the tag format's uint16 length prefix) still needs a runtime check — done in NewStringTag.
// A bare `StringTag("...")` conversion skips that check, for use only when the length is already
// known-valid (e.g. it just came off the wire, where the length prefix itself limits it).
type StringTag string

func NewStringTag(value string) (StringTag, error) {
	if len(value) > 32767 {
		return "", NewInvalidTagValueException(fmt.Sprintf("StringTag cannot hold more than 32767 bytes, got string of length %d", len(value)))
	}
	return StringTag(value), nil
}

func (t StringTag) Type() Type            { return TagString }
func (t StringTag) Equals(other Tag) bool { o, ok := other.(StringTag); return ok && o == t }
func (t StringTag) String() string        { return t.stringify(0) }
func (t StringTag) stringify(int) string  { return `"` + string(t) + `"` }
func (t StringTag) write(w StreamWriter) error {
	return w.WriteString(string(t))
}

// ByteArrayTag is a port of pocketmine\nbt\tag\ByteArrayTag.
//
// This is a slice, not a value type like the scalars above, so Equals does a content comparison
// and container tags (CompoundTag/ListTag) must copy the slice when cloning — see CloneTag.
type ByteArrayTag []byte

func (t ByteArrayTag) Type() Type { return TagByteArray }
func (t ByteArrayTag) Equals(other Tag) bool {
	o, ok := other.(ByteArrayTag)
	if !ok || len(o) != len(t) {
		return false
	}
	for i := range t {
		if t[i] != o[i] {
			return false
		}
	}
	return true
}
func (t ByteArrayTag) String() string       { return t.stringify(0) }
func (t ByteArrayTag) stringify(int) string { return "b64:" + base64Encode(t) }
func (t ByteArrayTag) write(w StreamWriter) error {
	w.WriteByteArray(t)
	return nil
}

// IntArrayTag is a port of pocketmine\nbt\tag\IntArrayTag.
type IntArrayTag []int32

func (t IntArrayTag) Type() Type { return TagIntArray }
func (t IntArrayTag) Equals(other Tag) bool {
	o, ok := other.(IntArrayTag)
	if !ok || len(o) != len(t) {
		return false
	}
	for i := range t {
		if t[i] != o[i] {
			return false
		}
	}
	return true
}
func (t IntArrayTag) String() string { return t.stringify(0) }
func (t IntArrayTag) stringify(int) string {
	s := "["
	for i, v := range t {
		if i > 0 {
			s += ","
		}
		s += strconv.FormatInt(int64(v), 10)
	}
	return s + "]"
}
func (t IntArrayTag) write(w StreamWriter) error {
	w.WriteIntArray(t)
	return nil
}
