package block

import "testing"

// TestBambooOffsetSeedMatchesReference cross-checks bambooOffsetSeed against values independently
// computed in Python using explicit 64-bit two's-complement wraparound at each step (the same
// semantics Go's int64 arithmetic guarantees), verifying the doc comment's claim that truncating
// to 64 bits along the way doesn't change the final 32-bit result.
func TestBambooOffsetSeedMatchesReference(t *testing.T) {
	cases := []struct {
		x, y, z int
		want    uint32
	}{
		{0, 0, 0, 0},
		{1, 0, 1, 2243233522},
		{5, 0, -3, 4076753538},
		{100, 0, 200, 3720487052},
		{-7, 0, 42, 2579149044},
	}
	for _, c := range cases {
		if got := bambooOffsetSeed(c.x, c.y, c.z); got != c.want {
			t.Errorf("bambooOffsetSeed(%d, %d, %d) = %d, want %d", c.x, c.y, c.z, got, c.want)
		}
	}
}

func newTestBambooSapling(w World) *BambooSapling {
	b := NewBambooSapling(mustBlockIdentifier(1038), "Test Bamboo Sapling", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBambooSaplingBreaksWithoutValidSupport(t *testing.T) {
	w := &noSupportWorld{}
	b := newTestBambooSapling(w)

	b.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

// replaceableBlock is a filler whose CanBeReplaced() is true, unlike testBlock/stemTestBlock
// which default to false (Block's default).
type replaceableBlock struct {
	Block
}

func (r *replaceableBlock) CanBeReplaced() bool { return true }

func (r *replaceableBlock) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

type replaceableWorld struct {
	fakeWorld
}

func (w *replaceableWorld) GetBlockAt(x, y, z int) Behavior {
	r := &replaceableBlock{Block: NewBlock(mustBlockIdentifier(1040), "Replaceable Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	r.Init(r)
	r.SetPosition(w, x, y, z)
	return r
}

func TestBambooSaplingBecomesReadyWhenSpaceAboveIsFree(t *testing.T) {
	w := &replaceableWorld{}
	b := newTestBambooSapling(w)

	b.OnRandomTick()

	if !b.Ready {
		t.Error("expected the sapling to become ready when the space above can be replaced")
	}
}

func TestBambooSeekToTopFindsTopOfStack(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}

	bottom := NewBamboo(mustBlockIdentifier(1039), "Test Bamboo Bottom", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	bottom.SetPosition(w, 1, 2, 3)
	middle := NewBamboo(mustBlockIdentifier(1039), "Test Bamboo Middle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	middle.SetPosition(w, 1, 3, 3)
	top := NewBamboo(mustBlockIdentifier(1039), "Test Bamboo Top", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	top.SetPosition(w, 1, 4, 3)

	w.blocks[[3]int{1, 2, 3}] = bottom
	w.blocks[[3]int{1, 3, 3}] = middle
	w.blocks[[3]int{1, 4, 3}] = top

	if got := bottom.seekToTop(); got != top {
		t.Error("seekToTop() did not find the topmost bamboo block")
	}
}

func newTestBambooForGrow(w World) *Bamboo {
	b := NewBamboo(mustBlockIdentifier(1039), "Test Bamboo", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func placeReplaceableAt(w *btWorld, x, y, z int) {
	r := &replaceableBlock{Block: NewBlock(mustBlockIdentifier(1040), "Replaceable Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	r.Init(r)
	r.SetPosition(w, x, y, z)
	w.blocks[[3]int{x, y, z}] = r
}

func TestBambooGrowFromHeightOneAddsOneSmallLeavesBlockAbove(t *testing.T) {
	w := newBTWorld()
	placeReplaceableAt(w, 1, 3, 3) // space above must be replaceable
	b := newTestBambooForGrow(w)

	if !b.grow(12, 1) {
		t.Fatal("expected grow to succeed")
	}
	if len(w.setCalls) != 1 {
		t.Fatalf("expected 1 SetBlock call, got %d", len(w.setCalls))
	}
	grown, ok := w.setCalls[0].blk.(*Bamboo)
	if !ok {
		t.Fatalf("expected a *Bamboo, got %T", w.setCalls[0].blk)
	}
	if grown.LeafSize != BambooSmallLeaves {
		t.Errorf("LeafSize = %d, want BambooSmallLeaves", grown.LeafSize)
	}
	if w.setCalls[0].pos.FloorY() != 3 {
		t.Errorf("new block Y = %d, want 3 (directly above the original)", w.setCalls[0].pos.FloorY())
	}
}

func TestBambooGrowFailsWhenSpaceAboveIsNotReplaceable(t *testing.T) {
	w := newBTWorld() // default filler above is not replaceable
	b := newTestBambooForGrow(w)

	if b.grow(12, 1) {
		t.Error("expected grow to fail without replaceable space above")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestBambooGrowFailsAtMaxHeight(t *testing.T) {
	w := newBTWorld()
	placeReplaceableAt(w, 1, 3, 3)
	// Stack bamboo below the growing block up to maxHeight so the height-counting loop bails out.
	below := NewBamboo(mustBlockIdentifier(1039), "Bamboo Below", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	below.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = below
	b := newTestBambooForGrow(w)

	if b.grow(2, 1) {
		t.Error("expected grow to fail once the stack reaches maxHeight")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}
