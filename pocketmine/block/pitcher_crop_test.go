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

func TestPitcherCropGrowAdvancesAgeBelowMax(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)
	p.Age = 0

	if !p.grow(nil) {
		t.Fatal("expected grow to report growth happened")
	}
	grown, ok := w.lastSetBlock.(*PitcherCrop)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *PitcherCrop, got %T", w.lastSetBlock)
	}
	if grown.Age != 1 {
		t.Errorf("grown.Age = %d, want 1", grown.Age)
	}
}

func TestPitcherCropGrowReturnsFalseAtMaxAge(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)
	p.Age = pitcherCropMaxAge

	if p.grow(nil) {
		t.Error("expected grow to return false at max age (turning into DoublePitcherCrop isn't ported)")
	}
	if w.lastSetBlock != nil {
		t.Error("expected no state change at max age")
	}
}

func TestPitcherCropOnInteractFertilizesWithBoneMeal(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)
	p.Age = 0
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if !p.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
}

func TestPitcherCropOnInteractIgnoresNonFertilizerItems(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)

	if p.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false for a non-fertilizer item")
	}
}

func TestPitcherCropOnInteractReturnsFalseAtMaxAgeEvenWithBoneMeal(t *testing.T) {
	w := &fakeWorld{}
	p := newTestPitcherCrop(w)
	p.Age = pitcherCropMaxAge
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if p.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false when grow() can't advance past max age")
	}
}
