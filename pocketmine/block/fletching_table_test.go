package block

import "testing"

func TestFletchingTableGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	f := NewFletchingTable(mustBlockIdentifier(1076), "Test Fletching Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)

	if f.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", f.GetFuelTime())
	}
}
