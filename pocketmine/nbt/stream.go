package nbt

// StreamReader is a port of the internal \NbtStreamReader interface.
//
// ReadByte/WriteByte deliberately match io.ByteReader/io.ByteWriter's exact signatures (byte,
// error) — not because this needs to satisfy those stdlib interfaces, but because `go vet`'s
// stdmethods check flags any type defining methods with these exact names against a different
// signature, and matching it is free (a byte-buffer append can't actually fail).
type StreamReader interface {
	ReadByte() (byte, error)
	ReadSignedByte() (int8, error)
	ReadShort() (uint16, error)
	ReadSignedShort() (int16, error)
	ReadInt() (int32, error)
	ReadLong() (int64, error)
	ReadFloat() (float32, error)
	ReadDouble() (float64, error)
	ReadByteArray() ([]byte, error)
	ReadString() (string, error)
	ReadIntArray() ([]int32, error)
}

// StreamWriter is a port of the internal \NbtStreamWriter interface.
type StreamWriter interface {
	WriteByte(v byte) error
	WriteShort(v uint16)
	WriteInt(v int32)
	WriteLong(v int64)
	WriteFloat(v float32)
	WriteDouble(v float64)
	WriteByteArray(v []byte)
	WriteString(v string) error
	WriteIntArray(v []int32)
}
