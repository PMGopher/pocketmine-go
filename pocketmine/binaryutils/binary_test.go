package binaryutils

import "testing"

func TestShortRoundTrip(t *testing.T) {
	b := WriteShort(0xABCD)
	v, err := ReadShort(b)
	if err != nil || v != 0xABCD {
		t.Fatalf("ReadShort(WriteShort(0xABCD)) = %v, %v", v, err)
	}
}

func TestIntSignExtension(t *testing.T) {
	b := WriteInt(-1)
	v, err := ReadInt(b)
	if err != nil || v != -1 {
		t.Fatalf("ReadInt(WriteInt(-1)) = %v, %v, want -1", v, err)
	}
}

func TestTriadRoundTrip(t *testing.T) {
	b := WriteTriad(0x123456)
	v, err := ReadTriad(b)
	if err != nil || v != 0x123456 {
		t.Fatalf("ReadTriad(WriteTriad(0x123456)) = %v, %v", v, err)
	}
	if len(b) != 3 {
		t.Fatalf("WriteTriad() length = %d, want 3", len(b))
	}
}

func TestFloatRoundTrip(t *testing.T) {
	b := WriteFloat(3.14)
	v, err := ReadFloat(b)
	if err != nil {
		t.Fatalf("ReadFloat() error = %v", err)
	}
	if v != float32(3.14) {
		t.Fatalf("ReadFloat(WriteFloat(3.14)) = %v, want 3.14", v)
	}
}

func TestVarIntZigzagRoundTrip(t *testing.T) {
	values := []int32{0, 1, -1, 2, -2, 127, -127, 128, -128, 2147483647, -2147483648}
	for _, want := range values {
		buf := WriteVarInt(want)
		offset := 0
		got, err := ReadVarInt(buf, &offset)
		if err != nil {
			t.Fatalf("ReadVarInt() error for %d: %v", want, err)
		}
		if got != want {
			t.Fatalf("ReadVarInt(WriteVarInt(%d)) = %d", want, got)
		}
		if offset != len(buf) {
			t.Fatalf("offset after ReadVarInt(%d) = %d, want %d", want, offset, len(buf))
		}
	}
}

func TestVarLongZigzagRoundTrip(t *testing.T) {
	values := []int64{0, 1, -1, 1000000000000, -1000000000000, 9223372036854775807, -9223372036854775808}
	for _, want := range values {
		buf := WriteVarLong(want)
		offset := 0
		got, err := ReadVarLong(buf, &offset)
		if err != nil {
			t.Fatalf("ReadVarLong() error for %d: %v", want, err)
		}
		if got != want {
			t.Fatalf("ReadVarLong(WriteVarLong(%d)) = %d", want, got)
		}
	}
}

func TestUnsignedVarIntMultiByte(t *testing.T) {
	// 300 requires 2 bytes in LEB128 (0xAC 0x02)
	buf := WriteUnsignedVarInt(300)
	if len(buf) != 2 {
		t.Fatalf("WriteUnsignedVarInt(300) length = %d, want 2", len(buf))
	}
	offset := 0
	got, err := ReadUnsignedVarInt(buf, &offset)
	if err != nil || got != 300 {
		t.Fatalf("ReadUnsignedVarInt() = %v, %v, want 300", got, err)
	}
}

func TestBinaryStreamReadWrite(t *testing.T) {
	s := NewBinaryStream(nil, 0)
	s.PutByte(42)
	s.PutShort(1000)
	s.PutVarInt(-12345)
	s.PutInt(-1)

	s.Rewind()
	b, err := s.GetByte()
	if err != nil || b != 42 {
		t.Fatalf("GetByte() = %v, %v, want 42", b, err)
	}
	sh, err := s.GetShort()
	if err != nil || sh != 1000 {
		t.Fatalf("GetShort() = %v, %v, want 1000", sh, err)
	}
	vi, err := s.GetVarInt()
	if err != nil || vi != -12345 {
		t.Fatalf("GetVarInt() = %v, %v, want -12345", vi, err)
	}
	i, err := s.GetInt()
	if err != nil || i != -1 {
		t.Fatalf("GetInt() = %v, %v, want -1", i, err)
	}
	if !s.Feof() {
		t.Fatalf("expected Feof() to be true after reading everything written")
	}
}

func TestNotEnoughBytesError(t *testing.T) {
	_, err := ReadInt([]byte{1, 2})
	if err == nil {
		t.Fatalf("expected an error reading a 4-byte int from 2 bytes")
	}
	if _, ok := err.(*BinaryDataError); !ok {
		t.Fatalf("expected *BinaryDataError, got %T", err)
	}
}
