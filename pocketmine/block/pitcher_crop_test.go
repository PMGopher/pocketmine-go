package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestPitcherCrop(w World) *PitcherCrop {
	p := NewPitcherCrop(mustBlockIdentifier(1063), "Test Pitcher Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	p.SetPosition(w, 1, 2, 3)
	return p
}

func TestPitcherCropCanBePlacedAtRequiresFarmland(t *testing.T) {
	farmland := &farmlandWorld{}
	p := newTestPitcherCrop(farmland)
	replace := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace.SetPosition(farmland, 1, 2, 3)
	if !p.CanBePlacedAt(replace, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to accept farmland below")
	}

	notFarmland := &candleWorld{}
	p2 := newTestPitcherCrop(notFarmland)
	replace2 := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace2.SetPosition(notFarmland, 1, 2, 3)
	if p2.CanBePlacedAt(replace2, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to reject non-farmland below")
	}
}

func TestPitcherCropOnNearbyBlockChangeBreaksWithoutFarmland(t *testing.T) {
	w := &cakeAirWorld{}
	p := newTestPitcherCrop(w)

	p.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestPitcherCropSetAgeRejectsOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)

	defer func() {
		if recover() == nil {
			t.Error("expected SetAge to panic for an out-of-range value")
		}
	}()
	p.SetAge(pitcherCropMaxAge + 1)
}

func TestPitcherCropCollisionBoxShrinksAfterAgeZero(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)

	young := p.RecalculateCollisionBoxes()
	p.SetAge(1)
	older := p.RecalculateCollisionBoxes()

	if len(young) != 1 || len(older) != 1 {
		t.Fatalf("expected exactly one collision box at any age")
	}
	if young[0] == older[0] {
		t.Error("expected the collision box to differ between age 0 and a later age")
	}
}
