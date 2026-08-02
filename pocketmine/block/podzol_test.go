package block

import "testing"

func newTestPodzol(w World) *Podzol {
	p := NewPodzol(mustBlockIdentifier(1093), "Test Podzol", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	p.SetPosition(w, 1, 2, 3)
	return p
}

func TestPodzolGetDropsForCompatibleToolReturnsDirt(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	p := newTestPodzol(w)

	drops := p.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != DIRT {
		t.Errorf("expected a Dirt drop, got %#v", drops[0])
	}
}
