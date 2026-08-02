package block

import "testing"

func newTestStone(w World) *Stone {
	s := NewStone(mustBlockIdentifier(STONE), "Test Stone", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestStoneGetDropsForCompatibleToolReturnsCobblestone(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	s := newTestStone(w)

	drops := s.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != COBBLESTONE {
		t.Errorf("expected a Cobblestone drop, got %#v", drops[0])
	}
}

func TestStoneIsAffectedBySilkTouch(t *testing.T) {
	w := &fakeWorld{}
	s := newTestStone(w)
	if !s.IsAffectedBySilkTouch() {
		t.Error("expected Stone to be affected by silk touch")
	}
}
