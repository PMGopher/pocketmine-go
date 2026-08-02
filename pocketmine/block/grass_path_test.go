package block

import "testing"

func newTestGrassPath(w World) *GrassPath {
	g := NewGrassPath(mustBlockIdentifier(1092), "Test Grass Path", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	g.SetPosition(w, 1, 2, 3)
	return g
}

func TestGrassPathOnNearbyBlockChangeBecomesDirtWhenCoveredBySolid(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	g := newTestGrassPath(w)

	solid := newTestBlock(false)
	solid.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = solid

	g.OnNearbyBlockChange()

	if _, ok := w.lastSetBlock.(*Dirt); !ok {
		t.Fatalf("expected SetBlock to be called with a *Dirt, got %T", w.lastSetBlock)
	}
}

func TestGrassPathOnNearbyBlockChangeStaysWhenUncovered(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	g := newTestGrassPath(w)

	air := VanillaAir().(*Air)
	air.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = air

	g.OnNearbyBlockChange()

	if w.lastSetBlock != nil {
		t.Errorf("expected no block change, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassPathGetDropsForCompatibleToolReturnsDirt(t *testing.T) {
	withFakeItemBlockFactory(t)
	w := &fakeWorld{}
	g := newTestGrassPath(w)

	drops := g.GetDropsForCompatibleTool(fakeItem{})
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	wrapped, ok := drops[0].(*fakeItemBlock)
	if !ok || wrapped.wrapped.GetTypeId() != DIRT {
		t.Errorf("expected a Dirt drop, got %#v", drops[0])
	}
}
