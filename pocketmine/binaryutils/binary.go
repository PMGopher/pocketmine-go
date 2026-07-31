package binaryutils

import (
	"encoding/binary"
	"math"
	"math/bits"
	"strconv"
)

// Binary is a port of pocketmine\utils\Binary.
//
// PHP has only one integer type (a 64-bit int), so the original leans on shift/mask tricks
// (`$value << 56 >> 56`, `& 0xff`, etc.) to simulate fixed-width signed/unsigned values. Go has
// real int8/16/32/64 and uint8/16/32/64 types, so most of those tricks become plain type
// conversions here instead.

func SignByte(v uint8) int8      { return int8(v) }
func UnsignByte(v int8) uint8    { return uint8(v) }
func SignShort(v uint16) int16   { return int16(v) }
func UnsignShort(v int16) uint16 { return uint16(v) }
func SignInt(v uint32) int32     { return int32(v) }
func UnsignInt(v int32) uint32   { return uint32(v) }

func FlipShortEndianness(v uint16) uint16 { return bits.ReverseBytes16(v) }
func FlipIntEndianness(v uint32) uint32   { return bits.ReverseBytes32(v) }
func FlipLongEndianness(v uint64) uint64  { return bits.ReverseBytes64(v) }

func need(b []byte, n int) error {
	if len(b) < n {
		return newBinaryDataError("Not enough bytes: need %d, have %d", n, len(b))
	}
	return nil
}

func ReadBool(b []byte) (bool, error) {
	if err := need(b, 1); err != nil {
		return false, err
	}
	return b[0] != 0x00, nil
}

func WriteBool(v bool) []byte {
	if v {
		return []byte{0x01}
	}
	return []byte{0x00}
}

// ReadByte reads an unsigned byte (0-255).
func ReadByte(b []byte) (uint8, error) {
	if err := need(b, 1); err != nil {
		return 0, err
	}
	return b[0], nil
}

// ReadSignedByte reads a signed byte (-128-127).
func ReadSignedByte(b []byte) (int8, error) {
	v, err := ReadByte(b)
	return int8(v), err
}

func WriteByte(v uint8) []byte { return []byte{v} }

func ReadShort(b []byte) (uint16, error) {
	if err := need(b, 2); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func ReadSignedShort(b []byte) (int16, error) {
	v, err := ReadShort(b)
	return int16(v), err
}

func WriteShort(v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
}

func ReadLShort(b []byte) (uint16, error) {
	if err := need(b, 2); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func ReadSignedLShort(b []byte) (int16, error) {
	v, err := ReadLShort(b)
	return int16(v), err
}

func WriteLShort(v uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return buf
}

func ReadTriad(b []byte) (uint32, error) {
	if err := need(b, 3); err != nil {
		return 0, err
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]), nil
}

func WriteTriad(v uint32) []byte {
	return []byte{byte(v >> 16), byte(v >> 8), byte(v)}
}

func ReadLTriad(b []byte) (uint32, error) {
	if err := need(b, 3); err != nil {
		return 0, err
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16, nil
}

func WriteLTriad(v uint32) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16)}
}

func ReadInt(b []byte) (int32, error) {
	if err := need(b, 4); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

func WriteInt(v int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v))
	return buf
}

func ReadLInt(b []byte) (int32, error) {
	if err := need(b, 4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func WriteLInt(v int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return buf
}

func ReadFloat(b []byte) (float32, error) {
	if err := need(b, 4); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b)), nil
}

func ReadRoundedFloat(b []byte, accuracy int) (float32, error) {
	v, err := ReadFloat(b)
	if err != nil {
		return 0, err
	}
	scale := float32(math.Pow(10, float64(accuracy)))
	return float32(math.Round(float64(v*scale))) / scale, nil
}

func WriteFloat(v float32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(v))
	return buf
}

func ReadLFloat(b []byte) (float32, error) {
	if err := need(b, 4); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b)), nil
}

func ReadRoundedLFloat(b []byte, accuracy int) (float32, error) {
	v, err := ReadLFloat(b)
	if err != nil {
		return 0, err
	}
	scale := float32(math.Pow(10, float64(accuracy)))
	return float32(math.Round(float64(v*scale))) / scale, nil
}

func WriteLFloat(v float32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
	return buf
}

// PrintFloat returns a printable representation of the float, with trailing zeros trimmed.
//
// The PHP original formats with sprintf("%F", ...) then regex-strips trailing zeros;
// strconv.FormatFloat's shortest-round-trip mode ('f', -1) already produces a clean decimal
// representation without trailing zeros, so no separate trimming step is needed here.
func PrintFloat(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}

func ReadDouble(b []byte) (float64, error) {
	if err := need(b, 8); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
}

func WriteDouble(v float64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

func ReadLDouble(b []byte) (float64, error) {
	if err := need(b, 8); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
}

func WriteLDouble(v float64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

func ReadLong(b []byte) (int64, error) {
	if err := need(b, 8); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

func WriteLong(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf
}

func ReadLLong(b []byte) (int64, error) {
	if err := need(b, 8); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func WriteLLong(v int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(v))
	return buf
}

// ReadUnsignedVarInt reads a 32-bit variable-length unsigned integer, advancing *offset.
func ReadUnsignedVarInt(buffer []byte, offset *int) (uint32, error) {
	var value uint32
	for i := 0; i <= 28; i += 7 {
		if *offset >= len(buffer) {
			return 0, newBinaryDataError("No bytes left in buffer")
		}
		b := buffer[*offset]
		*offset++
		value |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			return value, nil
		}
	}
	return 0, newBinaryDataError("VarInt did not terminate after 5 bytes!")
}

// ReadVarInt reads a 32-bit zigzag-encoded variable-length integer, advancing *offset.
func ReadVarInt(buffer []byte, offset *int) (int32, error) {
	raw, err := ReadUnsignedVarInt(buffer, offset)
	if err != nil {
		return 0, err
	}
	return int32(raw>>1) ^ -int32(raw&1), nil
}

// WriteUnsignedVarInt writes a 32-bit unsigned integer as a variable-length integer (up to 5
// bytes). Unlike the PHP original, this can never fail: a genuine uint32 always fits within 5
// groups of 7 bits, whereas PHP needed a defensive error path due to lacking a true 32-bit type.
func WriteUnsignedVarInt(value uint32) []byte {
	buf := make([]byte, 0, 5)
	for {
		bits := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			buf = append(buf, bits|0x80)
		} else {
			buf = append(buf, bits)
			return buf
		}
	}
}

// WriteVarInt writes a 32-bit integer as a zigzag-encoded variable-length integer.
func WriteVarInt(v int32) []byte {
	return WriteUnsignedVarInt(uint32((v << 1) ^ (v >> 31)))
}

// ReadUnsignedVarLong reads a 64-bit variable-length unsigned integer, advancing *offset.
func ReadUnsignedVarLong(buffer []byte, offset *int) (uint64, error) {
	var value uint64
	for i := 0; i <= 63; i += 7 {
		if *offset >= len(buffer) {
			return 0, newBinaryDataError("No bytes left in buffer")
		}
		b := buffer[*offset]
		*offset++
		value |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			return value, nil
		}
	}
	return 0, newBinaryDataError("VarLong did not terminate after 10 bytes!")
}

// ReadVarLong reads a 64-bit zigzag-encoded variable-length integer, advancing *offset.
func ReadVarLong(buffer []byte, offset *int) (int64, error) {
	raw, err := ReadUnsignedVarLong(buffer, offset)
	if err != nil {
		return 0, err
	}
	return int64(raw>>1) ^ -int64(raw&1), nil
}

// WriteUnsignedVarLong writes a 64-bit unsigned integer as a variable-length long (up to 10 bytes).
func WriteUnsignedVarLong(value uint64) []byte {
	buf := make([]byte, 0, 10)
	for {
		bits := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			buf = append(buf, bits|0x80)
		} else {
			buf = append(buf, bits)
			return buf
		}
	}
}

// WriteVarLong writes a 64-bit integer as a zigzag-encoded variable-length long.
func WriteVarLong(v int64) []byte {
	return WriteUnsignedVarLong(uint64((v << 1) ^ (v >> 63)))
}
