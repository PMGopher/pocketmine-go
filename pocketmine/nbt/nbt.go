package nbt

import "fmt"

// createTag is a port of NBT::createTag(): constructs the appropriate Tag for a type byte read
// off the wire.
func createTag(t Type, r StreamReader, tracker *ReaderTracker) (Tag, error) {
	switch t {
	case TagByte:
		v, err := r.ReadSignedByte()
		if err != nil {
			return nil, err
		}
		return ByteTag(v), nil
	case TagShort:
		v, err := r.ReadSignedShort()
		if err != nil {
			return nil, err
		}
		return ShortTag(v), nil
	case TagInt:
		v, err := r.ReadInt()
		if err != nil {
			return nil, err
		}
		return IntTag(v), nil
	case TagLong:
		v, err := r.ReadLong()
		if err != nil {
			return nil, err
		}
		return LongTag(v), nil
	case TagFloat:
		v, err := r.ReadFloat()
		if err != nil {
			return nil, err
		}
		return FloatTag(v), nil
	case TagDouble:
		v, err := r.ReadDouble()
		if err != nil {
			return nil, err
		}
		return DoubleTag(v), nil
	case TagByteArray:
		v, err := r.ReadByteArray()
		if err != nil {
			return nil, err
		}
		return ByteArrayTag(v), nil
	case TagString:
		v, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		return StringTag(v), nil
	case TagList:
		return readListTag(r, tracker)
	case TagCompound:
		return readCompoundTag(r, tracker)
	case TagIntArray:
		v, err := r.ReadIntArray()
		if err != nil {
			return nil, err
		}
		return IntArrayTag(v), nil
	default:
		return nil, NewNbtDataException(fmt.Sprintf("Unknown NBT tag type %d", t))
	}
}
