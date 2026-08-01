package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// dirtTaggedWorld returns a filler tagged BlockTypeTagsDirt for every GetBlockAt call.
type dirtTaggedWorld struct {
	fakeWorld
}

func (w *dirtTaggedWorld) GetBlockAt(x, y, z int) Behavior {
	filler := newTestBlock(false)
	filler.typeInfo = NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), []string{BlockTypeTagsDirt}, nil)
	filler.SetPosition(w, x, y, z)
	return filler
}

func newTestSapling(w World) *Sapling {
	s := NewSapling(mustBlockIdentifier(1060), "Test Sapling", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.SaplingTypeOak)
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestSaplingCanBePlacedAtRequiresDirtOrMud(t *testing.T) {
	dirt := &dirtTaggedWorld{}
	s := newTestSapling(dirt)
	replace := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace.SetPosition(dirt, 1, 2, 3)
	if !s.CanBePlacedAt(replace, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to accept a dirt-tagged block below")
	}

	notDirt := &candleWorld{}
	s2 := newTestSapling(notDirt)
	replace2 := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace2.SetPosition(notDirt, 1, 2, 3)
	if s2.CanBePlacedAt(replace2, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to reject a non-dirt block below")
	}
}

func TestSaplingOnNearbyBlockChangeBreaksWithoutSupport(t *testing.T) {
	w := &noSupportWorld{}
	s := newTestSapling(w)

	s.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestSaplingOnRandomTickBecomesReady(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSapling(w)

	for i := 0; i < 50 && !s.Ready; i++ {
		s.OnRandomTick()
	}
	if !s.Ready {
		t.Error("expected repeated OnRandomTick calls to eventually set Ready (full light, 1-in-7 chance)")
	}
}
