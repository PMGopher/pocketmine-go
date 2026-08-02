package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// grassWorld returns explicitly configured blocks/light levels for specific coordinates, and a
// transparent filler with full light (15) everywhere else.
type grassWorld struct {
	fakeWorld
	blocks    map[[3]int]Behavior
	fullLight map[[3]int]int
}

func (w *grassWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(true)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *grassWorld) GetFullLightAt(x, y, z int) int {
	if v, ok := w.fullLight[[3]int{x, y, z}]; ok {
		return v
	}
	return 15
}

func newTestGrass(w World) *Grass {
	g := NewGrass(mustBlockIdentifier(GRASS), "Test Grass", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	g.SetPosition(w, 1, 2, 3)
	return g
}

func newTestDirt(w World, x, y, z int, dirtType blockutils.DirtType) *Dirt {
	d := NewDirt(mustBlockIdentifier(1050), "Test Dirt", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	d.DirtTypeValue = dirtType
	d.SetPosition(w, x, y, z)
	return d
}

func TestGrassOnRandomTickDiesInLowLightUnderOpaqueBlock(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{{1, 3, 3}: 3}}
	opaqueAbove := newTestBlock(false)
	opaqueAbove.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = opaqueAbove
	g := newTestGrass(w)

	g.OnRandomTick()

	dirt, ok := w.lastSetBlock.(*Dirt)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *Dirt, got %T", w.lastSetBlock)
	}
	_ = dirt
}

func TestGrassOnRandomTickDoesNothingInMidLight(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{{1, 3, 3}: 5}}
	g := newTestGrass(w)

	g.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Errorf("expected no block change in mid light, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassOnRandomTickLowLightWithTransparentAboveDoesNotDie(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{{1, 3, 3}: 3}}
	g := newTestGrass(w)

	g.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Errorf("expected no death with transparent block above, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassTrySpreadOntoSpreadsOntoEligibleDirt(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeNormal)
	w.blocks[[3]int{5, 5, 5}] = dirt
	g := newTestGrass(w)

	g.trySpreadOnto(w, 5, 5, 5)

	grown, ok := w.lastSetBlock.(*Grass)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *Grass, got %T", w.lastSetBlock)
	}
	_ = grown
}

func TestGrassTrySpreadOntoSkipsNonDirtBlock(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	g := newTestGrass(w)

	g.trySpreadOnto(w, 5, 5, 5) // default filler is a testBlock, not *Dirt

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread onto a non-dirt block, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassTrySpreadOntoSkipsNonNormalDirtType(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeCoarse)
	w.blocks[[3]int{5, 5, 5}] = dirt
	g := newTestGrass(w)

	g.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread onto coarse dirt, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassTrySpreadOntoSkipsWhenLightAboveTooLow(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{{5, 6, 5}: 3}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeNormal)
	w.blocks[[3]int{5, 5, 5}] = dirt
	g := newTestGrass(w)

	g.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread with insufficient light above, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestGrassTrySpreadOntoSkipsWhenBlockAboveIsOpaque(t *testing.T) {
	w := &grassWorld{blocks: map[[3]int]Behavior{}, fullLight: map[[3]int]int{}}
	dirt := newTestDirt(w, 5, 5, 5, blockutils.DirtTypeNormal)
	w.blocks[[3]int{5, 5, 5}] = dirt
	opaqueAbove := newTestBlock(false)
	opaqueAbove.SetPosition(w, 5, 6, 5)
	w.blocks[[3]int{5, 6, 5}] = opaqueAbove
	g := newTestGrass(w)

	g.trySpreadOnto(w, 5, 5, 5)

	if w.lastSetBlock != nil {
		t.Errorf("expected no spread under an opaque block, got SetBlock(%T)", w.lastSetBlock)
	}
}
