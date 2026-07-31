package block

import "testing"

// testBlock is a minimal concrete Behavior used only to prove two things about the
// self-referencing Behavior pattern documented in behavior.go and block.go:
//
//  1. Block's own default methods that call other overridable methods (GetLightFilter ->
//     IsTransparent, BlocksDirectSkyLight -> GetLightFilter) actually reach the concrete
//     override through b.self, not Block's own (wrong) default.
//  2. Clone()+rebind produces an independent copy whose self points at the copy, not the
//     original — without rebind, the clone's self would still point at the pre-copy value.
type testBlock struct {
	Block
	transparent bool
}

func newTestBlock(transparent bool) *testBlock {
	idInfo, err := NewBlockIdentifier(999, nil)
	if err != nil {
		panic(err)
	}
	t := &testBlock{Block: NewBlock(idInfo, "Test Block", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	t.transparent = transparent
	t.Init(t)
	return t
}

func (t *testBlock) IsTransparent() bool { return t.transparent }

func (t *testBlock) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func TestSelfDispatchReachesOverride(t *testing.T) {
	opaque := newTestBlock(false)
	if opaque.GetLightFilter() != 15 {
		t.Errorf("opaque block: GetLightFilter() = %d, want 15", opaque.GetLightFilter())
	}
	if !opaque.BlocksDirectSkyLight() {
		t.Error("opaque block: BlocksDirectSkyLight() = false, want true")
	}

	transparent := newTestBlock(true)
	if transparent.GetLightFilter() != 0 {
		t.Errorf("transparent block: GetLightFilter() = %d, want 0", transparent.GetLightFilter())
	}
	if transparent.BlocksDirectSkyLight() {
		t.Error("transparent block: BlocksDirectSkyLight() = true, want false")
	}
}

func TestCloneRebindsIndependentSelf(t *testing.T) {
	original := newTestBlock(true)

	cloned := original.Clone().(*testBlock)
	cloned.transparent = false

	if !original.transparent {
		t.Fatal("mutating the clone's field mutated the original — Clone() is not producing an independent copy")
	}

	// If rebind hadn't repointed self at the clone, this would call original.IsTransparent()
	// (still true) instead of cloned.IsTransparent() (false), wrongly returning 0.
	if cloned.GetLightFilter() != 15 {
		t.Errorf("cloned.GetLightFilter() = %d, want 15 (clone's self is not correctly rebound)", cloned.GetLightFilter())
	}
	if original.GetLightFilter() != 0 {
		t.Errorf("original.GetLightFilter() = %d, want 0 (clone mutation leaked into original's self)", original.GetLightFilter())
	}
}
