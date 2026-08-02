package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// stemTestBlock is a minimal Behavior with a settable GetTypeId, used to stand in for a
// neighbouring Melon/Air block without needing the real block registry.
type stemTestBlock struct {
	Block
	typeID int
}

func (s *stemTestBlock) GetTypeId() int { return s.typeID }

func (s *stemTestBlock) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

// stemWorld returns a stemTestBlock with a configurable type ID for every GetBlockAt call, so
// Stem.OnNearbyBlockChange's neighbour check can be exercised without the block registry.
type stemWorld struct {
	fakeWorld
	neighborTypeID int
}

func (w *stemWorld) GetBlockAt(x, y, z int) Behavior {
	idInfo, err := NewBlockIdentifier(1009, nil)
	if err != nil {
		panic(err)
	}
	b := &stemTestBlock{Block: NewBlock(idInfo, "Stem Neighbor", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil)), typeID: w.neighborTypeID}
	b.Init(b)
	b.SetPosition(w, x, y, z)
	return b
}

func newTestMelonStem(w World) *MelonStem {
	idInfo, err := NewBlockIdentifier(1010, nil)
	if err != nil {
		panic(err)
	}
	m := NewMelonStem(idInfo, "Test Melon Stem", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	m.SetPosition(w, 1, 2, 3)
	return m
}

func TestMelonStemGetPlantTypeID(t *testing.T) {
	w := &stemWorld{}
	m := newTestMelonStem(w)
	if m.GetPlantTypeID() != MELON {
		t.Errorf("GetPlantTypeID() = %d, want MELON (%d)", m.GetPlantTypeID(), MELON)
	}
}

func TestStemResetsToUpWhenPointedMelonDisappears(t *testing.T) {
	w := &stemWorld{neighborTypeID: 0}
	m := newTestMelonStem(w)
	m.Facing = math.East

	m.OnNearbyBlockChange()

	if m.Facing != math.Up {
		t.Errorf("Facing = %v, want Up (should reset when the pointed-to block is no longer a melon)", m.Facing)
	}
}

func TestStemKeepsFacingWhenPointedMelonStillThere(t *testing.T) {
	w := &stemWorld{neighborTypeID: MELON}
	m := newTestMelonStem(w)
	m.Facing = math.East

	m.OnNearbyBlockChange()

	if m.Facing != math.East {
		t.Errorf("Facing = %v, want East (should stay pointed at the still-present melon)", m.Facing)
	}
}

func TestStemDefaultFacingIsUp(t *testing.T) {
	w := &stemWorld{}
	m := newTestMelonStem(w)
	if m.Facing != math.Up {
		t.Errorf("default Facing = %v, want Up", m.Facing)
	}
}

func TestMelonStemGetPlantReturnsAMelonBlock(t *testing.T) {
	w := &stemWorld{}
	m := newTestMelonStem(w)
	if got := m.GetPlant(); got.GetTypeId() != MELON {
		t.Errorf("GetPlant().GetTypeId() = %d, want MELON (%d)", got.GetTypeId(), MELON)
	}
}
