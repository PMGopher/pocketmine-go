package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestSmithingTableOnInteractReturnsTrue(t *testing.T) {
	w := &fakeWorld{}
	s := NewSmithingTable(mustBlockIdentifier(1077), "Test Smithing Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)

	if !s.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}

func TestSmithingTableGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	s := NewSmithingTable(mustBlockIdentifier(1077), "Test Smithing Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)

	if s.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", s.GetFuelTime())
	}
}
