package nbt

import (
	"pocketmine-go/pocketmine/binaryutils"
)

// endianCodec captures the only reads/writes that actually differ between big- and
// little-endian NBT (short/int/long/float/double/int-array); everything else (byte, byte array,
// string) is endian-agnostic and lives directly on Serializer.
//
// PHP expresses this the other way around: BaseNbtSerializer implements the shared logic and
// calls $this->readInt()/$this->writeShort() etc., relying on the BigEndian/LittleEndian
// subclass's override being picked up via normal virtual dispatch. Go has no such dispatch
// through an embedded struct, so the endian-specific half is pulled out into its own small
// interface and injected, rather than trying to fake inheritance.
type endianCodec interface {
	ReadShort(buf *binaryutils.BinaryStream) (uint16, error)
	ReadSignedShort(buf *binaryutils.BinaryStream) (int16, error)
	WriteShort(buf *binaryutils.BinaryStream, v uint16)
	ReadInt(buf *binaryutils.BinaryStream) (int32, error)
	WriteInt(buf *binaryutils.BinaryStream, v int32)
	ReadLong(buf *binaryutils.BinaryStream) (int64, error)
	WriteLong(buf *binaryutils.BinaryStream, v int64)
	ReadFloat(buf *binaryutils.BinaryStream) (float32, error)
	WriteFloat(buf *binaryutils.BinaryStream, v float32)
	ReadDouble(buf *binaryutils.BinaryStream) (float64, error)
	WriteDouble(buf *binaryutils.BinaryStream, v float64)
}

// Serializer is a port of pocketmine\nbt\BaseNbtSerializer, parameterized over endianness.
// Construct one with NewBigEndianSerializer or NewLittleEndianSerializer.
type Serializer struct {
	buffer *binaryutils.BinaryStream
	codec  endianCodec
}

func newSerializer(codec endianCodec) *Serializer {
	return &Serializer{buffer: binaryutils.NewBinaryStream(nil, 0), codec: codec}
}

func (s *Serializer) readRoot(maxDepth int) (*TreeRoot, error) {
	typeByte, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	if Type(typeByte) == TagEnd {
		return nil, NewNbtDataException("Found TAG_End at the start of buffer")
	}
	rootName, err := s.ReadString()
	if err != nil {
		return nil, err
	}
	tag, err := createTag(Type(typeByte), s, NewReaderTracker(maxDepth))
	if err != nil {
		return nil, err
	}
	return NewTreeRoot(tag, rootName)
}

// Read decodes NBT starting at offset in buffer, returning the root and the offset just past it.
func (s *Serializer) Read(buffer []byte, offset int, maxDepth int) (*TreeRoot, int, error) {
	s.buffer = binaryutils.NewBinaryStream(buffer, offset)
	root, err := s.readRoot(maxDepth)
	if err != nil {
		if _, ok := err.(*binaryutils.BinaryDataError); ok {
			return nil, 0, NewNbtDataException(err.Error())
		}
		return nil, 0, err
	}
	return root, s.buffer.GetOffset(), nil
}

// ReadHeadless reads a tag without a name/type header — the raw binary value of a tag whose type
// is already known by context. Used in a few places in the Bedrock network protocol.
func (s *Serializer) ReadHeadless(buffer []byte, rootType Type, offset int, maxDepth int) (Tag, int, error) {
	s.buffer = binaryutils.NewBinaryStream(buffer, offset)
	tag, err := createTag(rootType, s, NewReaderTracker(maxDepth))
	if err != nil {
		return nil, 0, err
	}
	return tag, s.buffer.GetOffset(), nil
}

// ReadMultiple decodes a sequence of back-to-back NBT blobs until the buffer is exhausted.
func (s *Serializer) ReadMultiple(buffer []byte, maxDepth int) ([]*TreeRoot, error) {
	s.buffer = binaryutils.NewBinaryStream(buffer, 0)
	var result []*TreeRoot
	for !s.buffer.Feof() {
		root, err := s.readRoot(maxDepth)
		if err != nil {
			if _, ok := err.(*binaryutils.BinaryDataError); ok {
				return nil, NewNbtDataException(err.Error())
			}
			return nil, err
		}
		result = append(result, root)
	}
	return result, nil
}

func (s *Serializer) writeRoot(root *TreeRoot) error {
	s.WriteByte(uint8(root.GetTag().Type()))
	if err := s.WriteString(root.GetName()); err != nil {
		return err
	}
	return root.GetTag().write(s)
}

func (s *Serializer) Write(data *TreeRoot) ([]byte, error) {
	s.buffer = binaryutils.NewBinaryStream(nil, 0)
	if err := s.writeRoot(data); err != nil {
		return nil, err
	}
	return s.buffer.GetBuffer(), nil
}

// WriteHeadless writes a nameless tag with no header. See ReadHeadless.
func (s *Serializer) WriteHeadless(data Tag) ([]byte, error) {
	s.buffer = binaryutils.NewBinaryStream(nil, 0)
	if err := data.write(s); err != nil {
		return nil, err
	}
	return s.buffer.GetBuffer(), nil
}

func (s *Serializer) WriteMultiple(data []*TreeRoot) ([]byte, error) {
	s.buffer = binaryutils.NewBinaryStream(nil, 0)
	for _, root := range data {
		if err := s.writeRoot(root); err != nil {
			return nil, err
		}
	}
	return s.buffer.GetBuffer(), nil
}

// --- shared, endian-agnostic reads/writes ---

func (s *Serializer) ReadByte() (byte, error) { return s.buffer.GetByte() }
func (s *Serializer) ReadSignedByte() (int8, error) {
	v, err := s.buffer.GetByte()
	return int8(v), err
}
func (s *Serializer) WriteByte(v byte) error { s.buffer.PutByte(v); return nil }

func (s *Serializer) ReadByteArray() ([]byte, error) {
	length, err := s.codec.ReadInt(s.buffer)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, NewNbtDataException("Array length cannot be less than zero")
	}
	return s.buffer.Get(int(length))
}

func (s *Serializer) WriteByteArray(v []byte) {
	s.codec.WriteInt(s.buffer, int32(len(v)))
	s.buffer.Put(v)
}

func checkReadStringLength(length int) (int, error) {
	if length > 32767 {
		return 0, NewNbtDataException("NBT string length too large")
	}
	return length, nil
}

func checkWriteStringLength(length int) (int, error) {
	if length > 32767 {
		return 0, NewInvalidTagValueException("NBT string length too large")
	}
	return length, nil
}

func (s *Serializer) ReadString() (string, error) {
	length, err := s.codec.ReadShort(s.buffer)
	if err != nil {
		return "", err
	}
	checkedLen, err := checkReadStringLength(int(length))
	if err != nil {
		return "", err
	}
	b, err := s.buffer.Get(checkedLen)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Serializer) WriteString(v string) error {
	checkedLen, err := checkWriteStringLength(len(v))
	if err != nil {
		return err
	}
	s.codec.WriteShort(s.buffer, uint16(checkedLen))
	s.buffer.Put([]byte(v))
	return nil
}

// --- endian-specific reads/writes, delegated to codec ---

func (s *Serializer) ReadShort() (uint16, error)      { return s.codec.ReadShort(s.buffer) }
func (s *Serializer) ReadSignedShort() (int16, error) { return s.codec.ReadSignedShort(s.buffer) }
func (s *Serializer) WriteShort(v uint16)             { s.codec.WriteShort(s.buffer, v) }
func (s *Serializer) ReadInt() (int32, error)         { return s.codec.ReadInt(s.buffer) }
func (s *Serializer) WriteInt(v int32)                { s.codec.WriteInt(s.buffer, v) }
func (s *Serializer) ReadLong() (int64, error)        { return s.codec.ReadLong(s.buffer) }
func (s *Serializer) WriteLong(v int64)               { s.codec.WriteLong(s.buffer, v) }
func (s *Serializer) ReadFloat() (float32, error)     { return s.codec.ReadFloat(s.buffer) }
func (s *Serializer) WriteFloat(v float32)            { s.codec.WriteFloat(s.buffer, v) }
func (s *Serializer) ReadDouble() (float64, error)    { return s.codec.ReadDouble(s.buffer) }
func (s *Serializer) WriteDouble(v float64)           { s.codec.WriteDouble(s.buffer, v) }

// ReadIntArray/WriteIntArray use plain signed 32-bit reads/writes (symmetric with IntTag).
//
// The PHP original reads with unpack("N*"/"V*") (always non-negative) but writes with
// pack("N*"/"V*") (which just truncates whatever int it's given mod 2^32) — so round-tripping a
// negative element through the real library silently turns it into a large positive one. That's
// a genuine asymmetry bug in the original, not an intentional format choice (the NBT spec's
// TAG_Int_Array is a signed 32-bit array, matching Java's `int[]`), so this port fixes it rather
// than reproducing it: every element round-trips exactly via ordinary sign-extending int32 I/O.
func (s *Serializer) ReadIntArray() ([]int32, error) {
	length, err := s.codec.ReadInt(s.buffer)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, NewNbtDataException("Array length cannot be less than zero")
	}
	result := make([]int32, length)
	for i := range result {
		v, err := s.codec.ReadInt(s.buffer)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

func (s *Serializer) WriteIntArray(array []int32) {
	s.codec.WriteInt(s.buffer, int32(len(array)))
	for _, v := range array {
		s.codec.WriteInt(s.buffer, v)
	}
}

var _ StreamReader = (*Serializer)(nil)
var _ StreamWriter = (*Serializer)(nil)
