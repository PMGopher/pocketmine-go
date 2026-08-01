package block

import "testing"

// sugarcaneWorld returns a fixed type ID for every GetBlockAt call except at "self" positions
// (where a real Sugarcane stack can be placed by the test).
type sugarcaneWorld struct {
	fakeWorld
	fillerTypeID int
	blocks       map[[3]int]Behavior
}

func (w *sugarcaneWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := &stemTestBlock{typeID: w.fillerTypeID}
	filler.Block = NewBlock(mustBlockIdentifier(1018), "Sugarcane Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	filler.Init(filler)
	filler.SetPosition(w, x, y, z)
	return filler
}

func mustBlockIdentifier(id int) *BlockIdentifier {
	idInfo, err := NewBlockIdentifier(id, nil)
	if err != nil {
		panic(err)
	}
	return idInfo
}

func newTestSugarcane(w World) *Sugarcane {
	s := NewSugarcane(mustBlockIdentifier(1019), "Test Sugarcane", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestSugarcaneBreaksWithoutSupportOrWater(t *testing.T) {
	w := &sugarcaneWorld{fillerTypeID: 0}
	s := newTestSugarcane(w)

	s.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d calls", len(w.breakCalls))
	}
}

func TestSugarcaneAgesUpWhenSupportedByNearbyWater(t *testing.T) {
	w := &sugarcaneWorld{fillerTypeID: 0}
	waterBlock := &stemTestBlock{typeID: WATER}
	waterBlock.Block = NewBlock(mustBlockIdentifier(1020), "Water Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	waterBlock.Init(waterBlock)
	waterBlock.SetPosition(w, 2, 1, 3) // East of (1,1,3), the block below the sugarcane.

	w.blocks = map[[3]int]Behavior{{2, 1, 3}: waterBlock}

	s := newTestSugarcane(w)
	s.OnRandomTick()

	if len(w.breakCalls) != 0 {
		t.Fatalf("expected no break when nearby water supports growth, got %d break calls", len(w.breakCalls))
	}
	if s.Age != 1 {
		t.Errorf("Age = %d, want 1", s.Age)
	}
}
