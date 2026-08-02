package block

import "testing"

func newTestBigDripleafStem(w World) *BigDripleafStem {
	b := NewBigDripleafStem(mustBlockIdentifier(1101), "Test Big Dripleaf Stem", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBigDripleafStemGetDropsForCompatibleToolReturnsHead(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	b := newTestBigDripleafStem(w)

	drops := b.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != BIG_DRIPLEAF_HEAD {
		t.Errorf("expected a Big Dripleaf Head drop, got %#v", drops[0])
	}
}

func TestBigDripleafStemGetPickedItemReturnsHead(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	b := newTestBigDripleafStem(w)

	picked := b.GetPickedItem(false)
	wrapped, ok := picked.(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != BIG_DRIPLEAF_HEAD {
		t.Errorf("expected a Big Dripleaf Head item, got %#v", picked)
	}
}
