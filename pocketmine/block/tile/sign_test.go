package tile

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestSignSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	s := NewSign(w, math.Vector3{})

	red := color.NewColor(255, 0, 0)
	s.SetText(blockutils.NewSignText([]string{"Hello", "World"}, &red, true))
	s.SetWaxed(true)

	saved := s.SaveNBT()

	decoded := NewSign(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}

	lines := decoded.GetText().GetLines()
	if lines[0] != "Hello" || lines[1] != "World" {
		t.Errorf("lines = %v, want [Hello World ...]", lines)
	}
	if !decoded.GetText().IsGlowing() {
		t.Error("expected glowing text to survive the round trip")
	}
	if decoded.GetText().GetBaseColor() != red {
		t.Errorf("base color = %v, want %v", decoded.GetText().GetBaseColor(), red)
	}
	if !decoded.IsWaxed() {
		t.Error("expected Waxed to survive the round trip")
	}
}

func TestSignReadSaveDataLegacyPerLineFormat(t *testing.T) {
	w := &fakeWorld{}
	tag := nbt.NewCompoundTag()
	tag.SetString("Text1", "line one")
	tag.SetString("Text3", "line three")

	s := NewSign(w, math.Vector3{})
	if err := s.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	lines := s.GetText().GetLines()
	if lines[0] != "line one" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "line one")
	}
	if lines[2] != "line three" {
		t.Errorf("lines[2] = %q, want %q", lines[2], "line three")
	}
}

func TestSignAddAdditionalSpawnDataIncludesEditorID(t *testing.T) {
	w := &fakeWorld{}
	s := NewSign(w, math.Vector3{})

	tag := nbt.NewCompoundTag()
	s.AddAdditionalSpawnData(tag)
	if got, err := tag.GetLong(SignTagLockedForEditing); err != nil || int64(got) != -1 {
		t.Errorf("locked-for-editing = %d, err %v, want -1 (no editor)", got, err)
	}

	s.SetEditorEntityRuntimeID(42, true)
	tag2 := nbt.NewCompoundTag()
	s.AddAdditionalSpawnData(tag2)
	if got, err := tag2.GetLong(SignTagLockedForEditing); err != nil || int64(got) != 42 {
		t.Errorf("locked-for-editing = %d, err %v, want 42", got, err)
	}
}
