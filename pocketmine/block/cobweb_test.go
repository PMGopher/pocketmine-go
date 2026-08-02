package block

import "testing"

func newTestCobweb(w World) *Cobweb {
	c := NewCobweb(mustBlockIdentifier(1102), "Test Cobweb", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCobwebGetDropsForCompatibleToolWithShearsDropsItself(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	c := newTestCobweb(w)

	drops := c.GetDropsForCompatibleTool(fakeItem{toolType: ToolTypeShears})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != c.GetTypeId() {
		t.Errorf("expected a Cobweb drop, got %#v", drops[0])
	}
}

func TestCobwebGetDropsForCompatibleToolWithoutShearsReturnsNil(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	c := newTestCobweb(w)

	if drops := c.GetDropsForCompatibleTool(fakeItem{toolType: ToolTypeShovel}); drops != nil {
		t.Errorf("expected nil drops without shears (VanillaItems.STRING() isn't ported), got %v", drops)
	}
}
