package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestCartographyTableOnInteractReturnsTrue(t *testing.T) {
	w := &fakeWorld{}
	c := NewCartographyTable(mustBlockIdentifier(1073), "Test Cartography Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	if !c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}

func TestCartographyTableGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	c := NewCartographyTable(mustBlockIdentifier(1073), "Test Cartography Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	if c.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", c.GetFuelTime())
	}
}
