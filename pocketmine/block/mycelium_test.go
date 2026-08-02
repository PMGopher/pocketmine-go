package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

func newTestMycelium(w World) *Mycelium {
	m := NewMycelium(mustBlockIdentifier(MYCELIUM), "Test Mycelium", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	m.SetPosition(w, 1, 2, 3)
	return m
}

func TestMyceliumTrySpreadOntoSpreadsOntoEligibleDirt(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeNormal)
	w.blocks[[3]int{5, 5, 5}] = dirt
	m := newTestMycelium(w)

	m.trySpreadOnto(w, 5, 5, 5)

	if _, ok := w.lastSetBlock.(*Mycelium); !ok {
		t.Fatalf("expected SetBlock to be called with a *Mycelium, got %T", w.lastSetBlock)
	}
}

func TestMyceliumTrySpreadOntoSkipsNonDirtBlock(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	m := newTestMycelium(w)

	m.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread onto a non-dirt block, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestMyceliumTrySpreadOntoSkipsNonNormalDirtType(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeRooted)
	w.blocks[[3]int{5, 5, 5}] = dirt
	m := newTestMycelium(w)

	m.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread onto rooted dirt, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestMyceliumTrySpreadOntoSkipsWhenBlockAboveIsOpaque(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeNormal)
	w.blocks[[3]int{5, 5, 5}] = dirt
	opaqueAbove := newTestBlock(false)
	opaqueAbove.SetPosition(w, 5, 6, 5)
	w.blocks[[3]int{5, 6, 5}] = opaqueAbove
	m := newTestMycelium(w)

	m.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread under an opaque block, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestMyceliumGetDropsForCompatibleToolReturnsDirt(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	m := newTestMycelium(w)

	drops := m.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != DIRT {
		t.Errorf("expected a Dirt drop, got %#v", drops[0])
	}
}
