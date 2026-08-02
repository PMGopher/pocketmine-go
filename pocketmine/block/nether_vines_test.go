package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// vineWorld stores blocks by coordinate, defaulting to a plain filler elsewhere.
type vineWorld struct {
	fakeWorld
	blocks     map[[3]int]Behavior
	outOfWorld map[[3]int]bool
	setCalls   []cactusSetCall
}

func (w *vineWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(false)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *vineWorld) IsInWorld(x, y, z int) bool { return !w.outOfWorld[[3]int{x, y, z}] }

func (w *vineWorld) SetBlock(pos Position, blk Behavior) error {
	w.setCalls = append(w.setCalls, cactusSetCall{pos, blk})
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
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

func TestNetherVinesGrowAddsOneSegmentAndIncrementsAge(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}, outOfWorld: map[[3]int]bool{}}
	n := newTestNetherVines(w, 1, 5, 3)
	n.Age = 3
	w.blocks[[3]int{1, 5, 3}] = n
	// Default filler below (growth direction Down) is opaque/non-replaceable... need replaceable.
	replaceable := &replaceableBlock{Block: NewBlock(mustBlockIdentifier(1040), "Replaceable", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	replaceable.Init(replaceable)
	replaceable.SetPosition(w, 1, 4, 3)
	w.blocks[[3]int{1, 4, 3}] = replaceable

	if !n.grow(1) {
		t.Fatal("expected grow to succeed")
	}
	if len(w.setCalls) != 1 {
		t.Fatalf("expected 1 SetBlock call, got %d", len(w.setCalls))
	}
	grown, ok := w.setCalls[0].blk.(*NetherVines)
	if !ok {
		t.Fatalf("expected a *NetherVines, got %T", w.setCalls[0].blk)
	}
	if grown.Age != 4 {
		t.Errorf("Age = %d, want 4", grown.Age)
	}
	if w.setCalls[0].pos.FloorY() != 4 {
		t.Errorf("new segment Y = %d, want 4 (one step down)", w.setCalls[0].pos.FloorY())
	}
}

func TestNetherVinesGrowCapsAgeAtMax(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}, outOfWorld: map[[3]int]bool{}}
	n := newTestNetherVines(w, 1, 5, 3)
	n.Age = NetherVinesMaxAge
	w.blocks[[3]int{1, 5, 3}] = n
	replaceable := &replaceableBlock{Block: NewBlock(mustBlockIdentifier(1040), "Replaceable", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	replaceable.Init(replaceable)
	replaceable.SetPosition(w, 1, 4, 3)
	w.blocks[[3]int{1, 4, 3}] = replaceable

	if !n.grow(1) {
		t.Fatal("expected grow to succeed")
	}
	grown := w.setCalls[0].blk.(*NetherVines)
	if grown.Age != NetherVinesMaxAge {
		t.Errorf("Age = %d, want capped at %d", grown.Age, NetherVinesMaxAge)
	}
}

func TestNetherVinesGrowFailsWhenNotReplaceable(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}, outOfWorld: map[[3]int]bool{}}
	n := newTestNetherVines(w, 1, 5, 3)
	w.blocks[[3]int{1, 5, 3}] = n
	// Default filler below is opaque/non-replaceable.

	if n.grow(1) {
		t.Error("expected grow to fail when the space below isn't replaceable")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}
