package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// vineWorld stores blocks by coordinate, defaulting to a plain filler elsewhere.
type vineWorld struct {
	fakeWorld
	blocks map[[3]int]Behavior
}

func (w *vineWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(false)
	filler.SetPosition(w, x, y, z)
	return filler
}

func newTestNetherVines(w World, x, y, z int) *NetherVines {
	n := NewNetherVines(mustBlockIdentifier(1028), "Test Nether Vines", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), math.Down)
	n.SetPosition(w, x, y, z)
	return n
}

func TestNetherVinesSeekToTipFindsBottomOfChain(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}

	top := newTestNetherVines(w, 1, 5, 3)
	middle := newTestNetherVines(w, 1, 4, 3)
	bottom := newTestNetherVines(w, 1, 3, 3)

	w.blocks[[3]int{1, 5, 3}] = top
	w.blocks[[3]int{1, 4, 3}] = middle
	w.blocks[[3]int{1, 3, 3}] = bottom

	tip := top.seekToTip()
	if tip != bottom {
		t.Errorf("seekToTip() found the wrong block (growth direction is Down, so the tip should be the bottom-most vine)")
	}
}

func TestNetherVinesSeekToTipStopsAtNonVineNeighbor(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	only := newTestNetherVines(w, 1, 5, 3)
	w.blocks[[3]int{1, 5, 3}] = only

	if tip := only.seekToTip(); tip != only {
		t.Error("expected seekToTip to return the vine itself when there's no further vine below")
	}
}
