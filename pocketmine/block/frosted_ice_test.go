package block

import "testing"

func newTestFrostedIce(w World) *FrostedIce {
	f := NewFrostedIce(mustBlockIdentifier(1094), "Test Frosted Ice", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)
	return f
}

func TestFrostedIceTryMeltBecomesWaterWhenFullyAged(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFrostedIce(w)
	f.Age = FrostedIceMaxAge

	if !f.tryMelt() {
		t.Fatal("expected tryMelt to report destroyed")
	}
	if _, ok := w.lastSetBlock.(*Water); !ok {
		t.Fatalf("expected SetBlock to be called with a *Water, got %T", w.lastSetBlock)
	}
}

func TestFrostedIceTryMeltAgesWithoutMeltingWhenNotFullyAged(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFrostedIce(w)
	f.Age = 0

	if f.tryMelt() {
		t.Error("expected tryMelt to report not-yet-destroyed")
	}
	if f.Age != 1 {
		t.Errorf("Age = %d, want 1", f.Age)
	}
	if _, ok := w.lastSetBlock.(*Water); ok {
		t.Error("expected no water swap while still aging")
	}
}
