package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestVineCloneCopiesFacesMapIndependently(t *testing.T) {
	idInfo, err := NewBlockIdentifier(1002, nil)
	if err != nil {
		t.Fatal(err)
	}
	original := NewVine(idInfo, "Test Vine", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	original.SetFace(math.North, true)

	clone := original.Clone().(*Vine)
	clone.SetFace(math.South, true)

	if original.HasFace(math.South) {
		t.Fatal("Clone() shares the same Faces map as the original — mutating the clone mutated it")
	}
	if !original.HasFace(math.North) {
		t.Fatal("original lost its own face after cloning")
	}
	if !clone.HasFace(math.North) {
		t.Fatal("clone did not inherit the original's faces at clone time")
	}
}
