package block

import "testing"

func TestUnknownBlockCanBePlacedIsFalse(t *testing.T) {
	w := &fakeWorld{}
	u := NewUnknownBlock(mustBlockIdentifier(1079), NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), 42)
	u.SetPosition(w, 1, 2, 3)

	if u.CanBePlaced() {
		t.Error("expected CanBePlaced() to be false")
	}
}

func TestUnknownBlockGetDropsIsEmpty(t *testing.T) {
	w := &fakeWorld{}
	u := NewUnknownBlock(mustBlockIdentifier(1079), NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), 42)
	u.SetPosition(w, 1, 2, 3)

	if drops := u.GetDrops(fakeItem{}); len(drops) != 0 {
		t.Errorf("GetDrops() = %v, want empty", drops)
	}
}

func TestUnknownBlockDescribeBlockItemStateRoundTripsStateData(t *testing.T) {
	w := &fakeWorld{}
	u := NewUnknownBlock(mustBlockIdentifier(1079), NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), 42)
	u.SetPosition(w, 1, 2, 3)

	if u.stateData != 42 {
		t.Errorf("stateData = %d, want 42", u.stateData)
	}
}
