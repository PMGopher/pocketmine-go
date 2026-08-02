package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestSmallDripleaf(w World, x, y, z int) *SmallDripleaf {
	s := NewSmallDripleaf(mustBlockIdentifier(1029), "Test Small Dripleaf", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, x, y, z)
	return s
}

func TestSmallDripleafGetAffectedBlocksIncludesOtherHalf(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}

	bottom := newTestSmallDripleaf(w, 1, 2, 3)
	top := newTestSmallDripleaf(w, 1, 3, 3)
	top.Top = true
	w.blocks[[3]int{1, 2, 3}] = bottom
	w.blocks[[3]int{1, 3, 3}] = top

	affected := bottom.GetAffectedBlocks()
	if len(affected) != 2 {
		t.Fatalf("GetAffectedBlocks() returned %d blocks, want 2", len(affected))
	}
	if affected[0] != Behavior(bottom) || affected[1] != Behavior(top) {
		t.Error("expected GetAffectedBlocks to return [bottom, top]")
	}
}

func TestSmallDripleafGetAffectedBlocksFallsBackWithoutOtherHalf(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	bottom := newTestSmallDripleaf(w, 1, 2, 3)
	w.blocks[[3]int{1, 2, 3}] = bottom

	affected := bottom.GetAffectedBlocks()
	if len(affected) != 1 || affected[0] != Behavior(bottom) {
		t.Errorf("expected GetAffectedBlocks to fall back to just itself, got %v", affected)
	}
}

func TestSmallDripleafBreaksWithoutClaySupport(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	filler := newTestBlock(false)
	filler.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = filler

	// The other-half check also needs to see itself there, so put a matching top half above.
	top := newTestSmallDripleaf(w, 1, 3, 3)
	top.Top = true
	w.blocks[[3]int{1, 3, 3}] = top

	bottom := newTestSmallDripleaf(w, 1, 2, 3)
	w.blocks[[3]int{1, 2, 3}] = bottom

	bottom.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once (no clay support), got %d", len(w.breakCalls))
	}
}

func TestSmallDripleafPlaceAddsTopHalfToTransaction(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	clay := &stemTestBlock{typeID: CLAY}
	clay.Block = NewBlock(mustBlockIdentifier(1030), "Clay Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	clay.Init(clay)
	clay.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = clay

	air := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	air.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = air

	replace := newTestSmallDripleaf(w, 1, 2, 3)
	bottom := newTestSmallDripleaf(w, 1, 2, 3)
	bottom.Facing = math.East

	tx := &recordingTransaction{}
	if !bottom.Place(tx, fakeItem{}, replace, replace, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed")
	}

	if len(tx.blocks) != 2 {
		t.Fatalf("expected 2 blocks added to the transaction (top half + the placed bottom), got %d", len(tx.blocks))
	}
	top, ok := tx.blocks[0].(*SmallDripleaf)
	if !ok || !top.Top {
		t.Fatalf("expected the first added block to be a top-half *SmallDripleaf, got %#v", tx.blocks[0])
	}
	if top.Facing != math.East {
		t.Errorf("top half Facing = %v, want East (carried from the bottom)", top.Facing)
	}
}

func TestSmallDripleafPlaceFailsWithoutSpaceAbove(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	clay := &stemTestBlock{typeID: CLAY}
	clay.Block = NewBlock(mustBlockIdentifier(1030), "Clay Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	clay.Init(clay)
	clay.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = clay
	// Default filler above (vineWorld) is opaque, not AIR.

	replace := newTestSmallDripleaf(w, 1, 2, 3)
	bottom := newTestSmallDripleaf(w, 1, 2, 3)

	tx := &recordingTransaction{}
	if bottom.Place(tx, fakeItem{}, replace, replace, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to fail without free space above")
	}
	if len(tx.blocks) != 0 {
		t.Errorf("expected no blocks added to the transaction, got %d", len(tx.blocks))
	}
}

func TestSmallDripleafSurvivesWithClaySupportAndOtherHalf(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}
	clay := &stemTestBlock{typeID: CLAY}
	clay.Block = NewBlock(mustBlockIdentifier(1030), "Clay Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	clay.Init(clay)
	clay.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = clay

	top := newTestSmallDripleaf(w, 1, 3, 3)
	top.Top = true
	w.blocks[[3]int{1, 3, 3}] = top

	bottom := newTestSmallDripleaf(w, 1, 2, 3)
	w.blocks[[3]int{1, 2, 3}] = bottom

	bottom.OnNearbyBlockChange()

	if len(w.breakCalls) != 0 {
		t.Errorf("expected no break with clay support and the top half present, got %d break calls", len(w.breakCalls))
	}
}
