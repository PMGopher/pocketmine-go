package nbt

import "pocketmine-go/pocketmine/binaryutils"

type bigEndianCodec struct{}

func (bigEndianCodec) ReadShort(buf *binaryutils.BinaryStream) (uint16, error) { return buf.GetShort() }
func (bigEndianCodec) ReadSignedShort(buf *binaryutils.BinaryStream) (int16, error) {
	return buf.GetSignedShort()
}
func (bigEndianCodec) WriteShort(buf *binaryutils.BinaryStream, v uint16)    { buf.PutShort(v) }
func (bigEndianCodec) ReadInt(buf *binaryutils.BinaryStream) (int32, error)  { return buf.GetInt() }
func (bigEndianCodec) WriteInt(buf *binaryutils.BinaryStream, v int32)       { buf.PutInt(v) }
func (bigEndianCodec) ReadLong(buf *binaryutils.BinaryStream) (int64, error) { return buf.GetLong() }
func (bigEndianCodec) WriteLong(buf *binaryutils.BinaryStream, v int64)      { buf.PutLong(v) }
func (bigEndianCodec) ReadFloat(buf *binaryutils.BinaryStream) (float32, error) {
	return buf.GetFloat()
}
func (bigEndianCodec) WriteFloat(buf *binaryutils.BinaryStream, v float32) { buf.PutFloat(v) }
func (bigEndianCodec) ReadDouble(buf *binaryutils.BinaryStream) (float64, error) {
	return buf.GetDouble()
}
func (bigEndianCodec) WriteDouble(buf *binaryutils.BinaryStream, v float64) { buf.PutDouble(v) }

// NewBigEndianSerializer is a port of pocketmine\nbt\BigEndianNbtSerializer — the "Java Edition"
// NBT byte order, also used for PC-format Bedrock world storage.
func NewBigEndianSerializer() *Serializer {
	return newSerializer(bigEndianCodec{})
}
