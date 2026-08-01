package block

import "testing"

func TestBrownMushroomGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrownMushroom(mustBlockIdentifier(1061), "Test Brown Mushroom", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)

	if b.GetLightLevel() != 1 {
		t.Errorf("GetLightLevel() = %d, want 1", b.GetLightLevel())
	}
}
