package nbt

import "pocketmine-go/pocketmine/binaryutils"

type littleEndianCodec struct{}

func (littleEndianCodec) ReadShort(buf *binaryutils.BinaryStream) (uint16, error) {
	return buf.GetLShort()
}
func (littleEndianCodec) ReadSignedShort(buf *binaryutils.BinaryStream) (int16, error) {
	return buf.GetSignedLShort()
}
func (littleEndianCodec) WriteShort(buf *binaryutils.BinaryStream, v uint16)   { buf.PutLShort(v) }
func (littleEndianCodec) ReadInt(buf *binaryutils.BinaryStream) (int32, error) { return buf.GetLInt() }
func (littleEndianCodec) WriteInt(buf *binaryutils.BinaryStream, v int32)      { buf.PutLInt(v) }
func (littleEndianCodec) ReadLong(buf *binaryutils.BinaryStream) (int64, error) {
	return buf.GetLLong()
}
func (littleEndianCodec) WriteLong(buf *binaryutils.BinaryStream, v int64) { buf.PutLLong(v) }
func (littleEndianCodec) ReadFloat(buf *binaryutils.BinaryStream) (float32, error) {
	return buf.GetLFloat()
}
func (littleEndianCodec) WriteFloat(buf *binaryutils.BinaryStream, v float32) { buf.PutLFloat(v) }
func (littleEndianCodec) ReadDouble(buf *binaryutils.BinaryStream) (float64, error) {
	return buf.GetLDouble()
}
func (littleEndianCodec) WriteDouble(buf *binaryutils.BinaryStream, v float64) { buf.PutLDouble(v) }

// NewLittleEndianSerializer is a port of pocketmine\nbt\LittleEndianNbtSerializer — used for
// Bedrock Edition's network protocol and its default world storage format.
func NewLittleEndianSerializer() *Serializer {
	return newSerializer(littleEndianCodec{})
}
