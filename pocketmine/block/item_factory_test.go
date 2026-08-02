package block

import "testing"

// fakeItemBlock is a minimal stand-in for item.ItemBlock, used to verify AsItem()/the
// GetDropsForCompatibleTool/GetSilkTouchDrops/GetPickedItem defaults route through
// NewItemBlockFunc correctly - without actually importing the item package (which block's own
// internal tests can never do: item already imports block, so the reverse would be a real Go
// import cycle).
type fakeItemBlock struct {
	fakeItem
	wrapped Behavior
}

// GetCount/SetCount shadow fakeItem's no-op stubs (fakeItem is embedded by value there since most
// callers don't care about count tracking) with real state, since drop-count-scaling code
// (EnderChest's SetCount(8), etc.) needs to see its SetCount call actually take effect.
func (f *fakeItemBlock) GetCount() int      { return f.count }
func (f *fakeItemBlock) SetCount(count int) { f.count = count }

func withFakeItemBlockFactory(t *testing.T) {
	t.Helper()
	old := NewItemBlockFunc
	NewItemBlockFunc = func(blk Behavior) Item {
		return &fakeItemBlock{fakeItem: fakeItem{typeID: -blk.GetTypeId(), count: 1}, wrapped: blk}
	}
	t.Cleanup(func() { NewItemBlockFunc = old })
}

func TestBlockAsItemErrorsWithoutItemFactory(t *testing.T) {
	NewItemBlockFunc = nil
	blk := newDropsTestBlock()

	if _, err := blk.AsItem(); err == nil {
		t.Error("expected AsItem to error when NewItemBlockFunc is nil")
	}
}

func TestBlockAsItemUsesFactoryWhenSet(t *testing.T) {
	withFakeItemBlockFactory(t)
	blk := newDropsTestBlock()

	got, err := blk.AsItem()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wrapped, ok := got.(*fakeItemBlock)
	if !ok {
		t.Fatalf("expected a *fakeItemBlock, got %T", got)
	}
	if wrapped.typeID != -blk.GetTypeId() {
		t.Errorf("item type ID = %d, want %d (negated block type ID)", wrapped.typeID, -blk.GetTypeId())
	}
}

func TestBlockGetDropsForCompatibleToolDefaultUsesAsItem(t *testing.T) {
	withFakeItemBlockFactory(t)
	blk := newTestBlock(true)

	drops := blk.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	if _, ok := drops[0].(*fakeItemBlock); !ok {
		t.Errorf("expected the drop to be a *fakeItemBlock, got %T", drops[0])
	}
}

func TestBlockGetDropsForCompatibleToolDefaultNilWithoutFactory(t *testing.T) {
	NewItemBlockFunc = nil
	blk := newTestBlock(true)

	if drops := blk.GetDropsForCompatibleTool(fakeItem{}); drops != nil {
		t.Errorf("expected nil drops without an item factory, got %v", drops)
	}
}

func TestBlockGetSilkTouchDropsDefaultUsesAsItem(t *testing.T) {
	withFakeItemBlockFactory(t)
	blk := newTestBlock(true)

	drops := blk.GetSilkTouchDrops(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
}

func TestBlockGetPickedItemDefaultUsesAsItem(t *testing.T) {
	withFakeItemBlockFactory(t)
	blk := newTestBlock(true)

	picked := blk.GetPickedItem(false)
	if _, ok := picked.(*fakeItemBlock); !ok {
		t.Errorf("expected a *fakeItemBlock, got %T", picked)
	}
}

func TestBlockGetPickedItemDefaultNilWithoutFactory(t *testing.T) {
	NewItemBlockFunc = nil
	blk := newTestBlock(true)

	if picked := blk.GetPickedItem(false); picked != nil {
		t.Errorf("expected nil without an item factory, got %v", picked)
	}
}
