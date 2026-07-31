package binaryutils

import "fmt"

// BinaryStream is a port of pocketmine\utils\BinaryStream: a cursor over a byte buffer with
// typed read/write helpers.
type BinaryStream struct {
	buffer []byte
	offset int
}

func NewBinaryStream(buffer []byte, offset int) *BinaryStream {
	return &BinaryStream{buffer: buffer, offset: offset}
}

func (s *BinaryStream) Rewind()              { s.offset = 0 }
func (s *BinaryStream) SetOffset(offset int) { s.offset = offset }
func (s *BinaryStream) GetOffset() int       { return s.offset }
func (s *BinaryStream) GetBuffer() []byte    { return s.buffer }

// Get reads len bytes from the buffer, advancing the offset.
func (s *BinaryStream) Get(length int) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	if length < 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	remaining := len(s.buffer) - s.offset
	if remaining < length {
		return nil, newBinaryDataError("Not enough bytes left in buffer: need %d, have %d", length, remaining)
	}
	result := s.buffer[s.offset : s.offset+length]
	s.offset += length
	return result, nil
}

// GetRemaining reads everything left in the buffer.
func (s *BinaryStream) GetRemaining() ([]byte, error) {
	if s.offset >= len(s.buffer) {
		return nil, newBinaryDataError("No bytes left to read")
	}
	result := s.buffer[s.offset:]
	s.offset = len(s.buffer)
	return result, nil
}

func (s *BinaryStream) Put(b []byte) { s.buffer = append(s.buffer, b...) }

func (s *BinaryStream) GetBool() (bool, error) {
	b, err := s.Get(1)
	if err != nil {
		return false, err
	}
	return b[0] != 0x00, nil
}

func (s *BinaryStream) PutBool(v bool) {
	if v {
		s.Put([]byte{0x01})
	} else {
		s.Put([]byte{0x00})
	}
}

func (s *BinaryStream) GetByte() (uint8, error) {
	b, err := s.Get(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (s *BinaryStream) PutByte(v uint8) { s.Put([]byte{v}) }

func (s *BinaryStream) GetShort() (uint16, error) {
	b, err := s.Get(2)
	if err != nil {
		return 0, err
	}
	return ReadShort(b)
}
func (s *BinaryStream) GetSignedShort() (int16, error) {
	v, err := s.GetShort()
	return int16(v), err
}
func (s *BinaryStream) PutShort(v uint16) { s.Put(WriteShort(v)) }

func (s *BinaryStream) GetLShort() (uint16, error) {
	b, err := s.Get(2)
	if err != nil {
		return 0, err
	}
	return ReadLShort(b)
}
func (s *BinaryStream) GetSignedLShort() (int16, error) {
	v, err := s.GetLShort()
	return int16(v), err
}
func (s *BinaryStream) PutLShort(v uint16) { s.Put(WriteLShort(v)) }

func (s *BinaryStream) GetTriad() (uint32, error) {
	b, err := s.Get(3)
	if err != nil {
		return 0, err
	}
	return ReadTriad(b)
}
func (s *BinaryStream) PutTriad(v uint32) { s.Put(WriteTriad(v)) }

func (s *BinaryStream) GetLTriad() (uint32, error) {
	b, err := s.Get(3)
	if err != nil {
		return 0, err
	}
	return ReadLTriad(b)
}
func (s *BinaryStream) PutLTriad(v uint32) { s.Put(WriteLTriad(v)) }

func (s *BinaryStream) GetInt() (int32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadInt(b)
}
func (s *BinaryStream) PutInt(v int32) { s.Put(WriteInt(v)) }

func (s *BinaryStream) GetLInt() (int32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadLInt(b)
}
func (s *BinaryStream) PutLInt(v int32) { s.Put(WriteLInt(v)) }

func (s *BinaryStream) GetFloat() (float32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadFloat(b)
}
func (s *BinaryStream) GetRoundedFloat(accuracy int) (float32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadRoundedFloat(b, accuracy)
}
func (s *BinaryStream) PutFloat(v float32) { s.Put(WriteFloat(v)) }

func (s *BinaryStream) GetLFloat() (float32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadLFloat(b)
}
func (s *BinaryStream) GetRoundedLFloat(accuracy int) (float32, error) {
	b, err := s.Get(4)
	if err != nil {
		return 0, err
	}
	return ReadRoundedLFloat(b, accuracy)
}
func (s *BinaryStream) PutLFloat(v float32) { s.Put(WriteLFloat(v)) }

func (s *BinaryStream) GetDouble() (float64, error) {
	b, err := s.Get(8)
	if err != nil {
		return 0, err
	}
	return ReadDouble(b)
}
func (s *BinaryStream) PutDouble(v float64) { s.Put(WriteDouble(v)) }

func (s *BinaryStream) GetLDouble() (float64, error) {
	b, err := s.Get(8)
	if err != nil {
		return 0, err
	}
	return ReadLDouble(b)
}
func (s *BinaryStream) PutLDouble(v float64) { s.Put(WriteLDouble(v)) }

func (s *BinaryStream) GetLong() (int64, error) {
	b, err := s.Get(8)
	if err != nil {
		return 0, err
	}
	return ReadLong(b)
}
func (s *BinaryStream) PutLong(v int64) { s.Put(WriteLong(v)) }

func (s *BinaryStream) GetLLong() (int64, error) {
	b, err := s.Get(8)
	if err != nil {
		return 0, err
	}
	return ReadLLong(b)
}
func (s *BinaryStream) PutLLong(v int64) { s.Put(WriteLLong(v)) }

func (s *BinaryStream) GetUnsignedVarInt() (uint32, error) {
	return ReadUnsignedVarInt(s.buffer, &s.offset)
}
func (s *BinaryStream) PutUnsignedVarInt(v uint32) { s.Put(WriteUnsignedVarInt(v)) }

func (s *BinaryStream) GetVarInt() (int32, error) {
	return ReadVarInt(s.buffer, &s.offset)
}
func (s *BinaryStream) PutVarInt(v int32) { s.Put(WriteVarInt(v)) }

func (s *BinaryStream) GetUnsignedVarLong() (uint64, error) {
	return ReadUnsignedVarLong(s.buffer, &s.offset)
}
func (s *BinaryStream) PutUnsignedVarLong(v uint64) { s.Put(WriteUnsignedVarLong(v)) }

func (s *BinaryStream) GetVarLong() (int64, error) {
	return ReadVarLong(s.buffer, &s.offset)
}
func (s *BinaryStream) PutVarLong(v int64) { s.Put(WriteVarLong(v)) }

// Feof returns whether the offset has reached the end of the buffer.
func (s *BinaryStream) Feof() bool {
	return s.offset >= len(s.buffer)
}
