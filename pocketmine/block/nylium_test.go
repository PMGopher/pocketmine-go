package block

import "testing"

func newTestNylium(w World) *Nylium {
	n := NewNylium(mustBlockIdentifier(CRIMSON_NYLIUM), "Test Nylium", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), nil)
	n.SetPosition(w, 1, 2, 3)
	return n
}

func newNetherrackAt(w World, x, y, z int) *stemTestBlock {
	b := &stemTestBlock{typeID: NETHERRACK}
	b.Block = NewBlock(mustBlockIdentifier(1051), "Netherrack Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.Init(b)
	b.SetPosition(w, x, y, z)
	return b
}

func TestNyliumOnRandomTickRevertsToNetherrackWhenCoveredByOpaque(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	opaqueAbove := newTestBlock(false)
	opaqueAbove.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = opaqueAbove
	n := newTestNylium(w)

	n.OnRandomTick()

	if _, ok := w.lastSetBlock.(*Netherrack); !ok {
		t.Fatalf("expected SetBlock to be called with a *Netherrack, got %T", w.lastSetBlock)
	}
}

func TestNyliumOnRandomTickDoesNotRevertWhenAboveIsTransparent(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	n := newTestNylium(w)

	n.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Errorf("expected no revert with transparent block above, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestNyliumTryConvertNetherrackAtConvertsEligibleNetherrack(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	w.blocks[[3]int{5, 5, 5}] = newNetherrackAt(w, 5, 5, 5)
	n := newTestNylium(w)

	n.tryConvertNetherrackAt(w, 5, 5, 5)

	if w.lastSetBlock == nil {
		t.Fatal("expected SetBlock to be called")
	}
	if w.lastSetBlock.GetTypeId() != CRIMSON_NYLIUM {
		t.Errorf("expected the replacement to keep the nylium's own type ID, got %d", w.lastSetBlock.GetTypeId())
	}
}

func TestNyliumTryConvertNetherrackAtSkipsNonNetherrack(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	n := newTestNylium(w)

	n.tryConvertNetherrackAt(w, 5, 5, 5) // default filler isn't Netherrack-typed

	if w.lastSetBlock != nil {
		t.Errorf("expected no conversion of a non-netherrack block, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestNyliumTryConvertNetherrackAtSkipsWhenAboveIsOpaque(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	w.blocks[[3]int{5, 5, 5}] = newNetherrackAt(w, 5, 5, 5)
	opaqueAbove := newTestBlock(false)
	opaqueAbove.SetPosition(w, 5, 6, 5)
	w.blocks[[3]int{5, 6, 5}] = opaqueAbove
	n := newTestNylium(w)

	n.tryConvertNetherrackAt(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no conversion under an opaque block, got SetBlock(%T)", w.lastSetBlock)
	}
}
