package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// neighborWorld returns a fixed neighbor Behavior for one specific coordinate, and an Air-like
// filler (no support) everywhere else.
type neighborWorld struct {
	fakeWorld
	at       [3]int
	neighbor Behavior
}

func (w *neighborWorld) GetBlockAt(x, y, z int) Behavior {
	if [3]int{x, y, z} == w.at {
		return w.neighbor
	}
	filler := newTestBlock(true)
	filler.SetPosition(w, x, y, z)
	return filler
}

func newTestWoodenFence(w World) *WoodenFence {
	idInfo, err := NewBlockIdentifier(1012, nil)
	if err != nil {
		panic(err)
	}
	f := NewWoodenFence(idInfo, "Test Wooden Fence", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	f.SetPosition(w, 1, 2, 3)
	return f
}

// TestWoodenFenceConnectsToNeighboringWoodenFence is a regression test: Fence.ReadStateFromWorld
// used to type-assert a neighbour to the exact *Fence type, which meant two WoodenFence blocks
// (which embed Fence, not are Fence) never recognised each other as connectable. Fixed via the
// fenceBaser interface.
func TestWoodenFenceConnectsToNeighboringWoodenFence(t *testing.T) {
	w := &neighborWorld{at: [3]int{2, 2, 3}}
	neighbor := newTestWoodenFence(w)
	w.neighbor = neighbor

	f := newTestWoodenFence(w)
	f.ReadStateFromWorld()

	if !f.Connections[math.East] {
		t.Error("expected a WoodenFence to connect to a neighbouring WoodenFence")
	}
}

func TestHardenedGlassPaneConnectsToNeighboringGlassPane(t *testing.T) {
	w := &neighborWorld{at: [3]int{2, 2, 3}}
	idInfo, err := NewBlockIdentifier(1013, nil)
	if err != nil {
		t.Fatal(err)
	}
	neighbor := NewGlassPane(idInfo, "Test Glass Pane", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	neighbor.SetPosition(w, 2, 2, 3)
	w.neighbor = neighbor

	idInfo2, err := NewBlockIdentifier(1014, nil)
	if err != nil {
		t.Fatal(err)
	}
	pane := NewHardenedGlassPane(idInfo2, "Test Hardened Glass Pane", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	pane.SetPosition(w, 1, 2, 3)
	pane.ReadStateFromWorld()

	if !pane.Connections[math.East] {
		t.Error("expected a HardenedGlassPane to connect to a neighbouring GlassPane")
	}
}
