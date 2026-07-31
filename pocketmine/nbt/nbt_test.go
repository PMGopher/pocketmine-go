package nbt

import "testing"

func buildSampleCompound(t *testing.T) *CompoundTag {
	t.Helper()
	str, err := NewStringTag("hello world")
	if err != nil {
		t.Fatalf("NewStringTag() error = %v", err)
	}
	list, err := NewListTag([]Tag{IntTag(1), IntTag(2), IntTag(3)}, TagInt)
	if err != nil {
		t.Fatalf("NewListTag() error = %v", err)
	}
	nested := NewCompoundTag().SetByte("flag", ByteTag(1))

	root := NewCompoundTag()
	root.SetByte("byte", ByteTag(-5))
	root.SetShort("short", ShortTag(-1000))
	root.SetInt("int", IntTag(123456))
	root.SetLong("long", LongTag(-123456789012345))
	root.SetFloat("float", FloatTag(3.5))
	root.SetDouble("double", DoubleTag(2.71828))
	root.SetByteArray("bytes", ByteArrayTag{1, 2, 3, 4})
	root.SetTag("string", str)
	root.SetTag("list", list)
	root.SetTag("nested", nested)
	root.SetIntArray("ints", IntArrayTag{-1, 0, 1, 2147483647, -2147483648})
	return root
}

func TestCompoundTagRoundTripBigEndian(t *testing.T) {
	root := buildSampleCompound(t)
	treeRoot, err := NewTreeRoot(root, "test")
	if err != nil {
		t.Fatalf("NewTreeRoot() error = %v", err)
	}

	encoded, err := NewBigEndianSerializer().Write(treeRoot)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	decoded, _, err := NewBigEndianSerializer().Read(encoded, 0, 512)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if !decoded.Equals(treeRoot) {
		t.Fatalf("decoded tree does not equal original\noriginal: %s\ndecoded:  %s", treeRoot, decoded)
	}
}

func TestCompoundTagRoundTripLittleEndian(t *testing.T) {
	root := buildSampleCompound(t)
	treeRoot, err := NewTreeRoot(root, "")
	if err != nil {
		t.Fatalf("NewTreeRoot() error = %v", err)
	}

	encoded, err := NewLittleEndianSerializer().Write(treeRoot)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	decoded, offset, err := NewLittleEndianSerializer().Read(encoded, 0, 512)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if offset != len(encoded) {
		t.Fatalf("offset = %d, want %d (whole buffer consumed)", offset, len(encoded))
	}

	if !decoded.Equals(treeRoot) {
		t.Fatalf("decoded tree does not equal original\noriginal: %s\ndecoded:  %s", treeRoot, decoded)
	}
}

func TestCompoundTagCloneIsIndependent(t *testing.T) {
	root := NewCompoundTag().SetByte("a", ByteTag(1))
	clone := root.Clone()
	clone.SetByte("a", ByteTag(2))

	orig, _ := root.GetByte("a")
	cloned, _ := clone.GetByte("a")
	if orig != 1 || cloned != 2 {
		t.Fatalf("Clone() was not independent: orig=%v cloned=%v", orig, cloned)
	}
}

func TestListTagRejectsMismatchedTagType(t *testing.T) {
	list, err := NewListTag([]Tag{IntTag(1)}, TagInt)
	if err != nil {
		t.Fatalf("NewListTag() error = %v", err)
	}
	if err := list.Push(StringTag("oops")); err == nil {
		t.Fatalf("expected an error pushing a StringTag onto a TagInt list")
	}
}

func TestGetTagValueErrors(t *testing.T) {
	c := NewCompoundTag().SetString("name", StringTag("Steve"))

	if _, err := c.GetByte("missing"); err == nil {
		t.Fatalf("expected NoSuchTagException for a missing tag")
	}
	if _, err := c.GetByte("name"); err == nil {
		t.Fatalf("expected UnexpectedTagTypeException for a type mismatch")
	}
	if got := c.GetByteOr("missing", ByteTag(42)); got != 42 {
		t.Fatalf("GetByteOr() = %v, want 42 (default)", got)
	}
}

func TestFloatTagEqualsUsesFloat32Precision(t *testing.T) {
	a := FloatTag(0.3)
	b := FloatTag(0.3)
	if !a.Equals(b) {
		t.Fatalf("expected two identical float32 values to be equal")
	}
}

func TestMergeOverwritesAndKeepsBoth(t *testing.T) {
	base := NewCompoundTag().SetInt("a", 1).SetInt("b", 1)
	other := NewCompoundTag().SetInt("b", 2).SetInt("c", 3)

	merged := base.Merge(other)
	a, _ := merged.GetInt("a")
	b, _ := merged.GetInt("b")
	c, _ := merged.GetInt("c")
	if a != 1 || b != 2 || c != 3 {
		t.Fatalf("Merge() = a=%v b=%v c=%v, want a=1 b=2 c=3", a, b, c)
	}
	// base itself must be untouched
	if origB, _ := base.GetInt("b"); origB != 1 {
		t.Fatalf("Merge() mutated the receiver: b=%v, want 1", origB)
	}
}
