package block

import "testing"

func newTestIce(w World) *Ice {
	i := NewIce(mustBlockIdentifier(1093), "Test Ice", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	i.SetPosition(w, 1, 2, 3)
	return i
}

func TestIceOnBreakMeltsForSurvivalPlayer(t *testing.T) {
	w := &fakeWorld{}
	i := newTestIce(w)

	if !i.OnBreak(fakeItem{}, &fakeSignPlayer{}, nil) {
		t.Fatal("expected OnBreak to return true")
	}
	if _, ok := w.lastSetBlock.(*Water); !ok {
		t.Fatalf("expected SetBlock to be called with a *Water, got %T", w.lastSetBlock)
	}
}

func TestIceOnBreakMeltsWithNoPlayer(t *testing.T) {
	w := &fakeWorld{}
	i := newTestIce(w)

	if !i.OnBreak(fakeItem{}, nil, nil) {
		t.Fatal("expected OnBreak to return true")
	}
	if _, ok := w.lastSetBlock.(*Water); !ok {
		t.Fatalf("expected SetBlock to be called with a *Water, got %T", w.lastSetBlock)
	}
}

// lowLightWorld reports adjacent light below the melt threshold.
type lowLightWorld struct {
	fakeWorld
}

func (w *lowLightWorld) GetHighestAdjacentBlockLightAt(x, y, z int) int { return 5 }

func TestIceOnRandomTickMeltsWhenLightIsHigh(t *testing.T) {
	w := &fakeWorld{} // GetHighestAdjacentBlockLightAt returns 15 by default
	i := newTestIce(w)

	i.OnRandomTick()

	if _, ok := w.lastSetBlock.(*Water); !ok {
		t.Fatalf("expected SetBlock to be called with a *Water, got %T", w.lastSetBlock)
	}
}

func TestIceOnRandomTickDoesNothingWhenLightIsLow(t *testing.T) {
	w := &lowLightWorld{}
	i := newTestIce(w)

	i.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Errorf("expected no melt, got SetBlock(%T)", w.lastSetBlock)
	}
}
