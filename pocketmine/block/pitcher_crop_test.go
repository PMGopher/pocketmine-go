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

// pitcherCropGrowWorld returns AIR everywhere by default (so the space above a max-age pitcher
// crop is free to grow into) and records every SetBlock call, unlike fakeWorld which only
// remembers the last one and returns nil from GetBlockAt (which would panic once grow() started
// reading the block above).
type pitcherCropGrowWorld struct {
	fakeWorld
	blocks   map[[3]int]Behavior
	setCalls []cactusSetCall
}

func (w *pitcherCropGrowWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	air := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	air.SetPosition(w, x, y, z)
	return air
}

func (w *pitcherCropGrowWorld) SetBlock(pos Position, blk Behavior) error {
	w.setCalls = append(w.setCalls, cactusSetCall{pos, blk})
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
}

func TestPitcherCropGrowAtMaxAgeBecomesDoublePitcherCropWhenSpaceAboveIsFree(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	p := newTestPitcherCrop(w)
	p.Age = pitcherCropMaxAge

	if !p.grow(nil) {
		t.Fatal("expected grow to succeed at max age with free space above")
	}
	if len(w.setCalls) != 2 {
		t.Fatalf("expected 2 SetBlock calls (bottom + top), got %d", len(w.setCalls))
	}
	bottom, ok := w.setCalls[0].blk.(*DoublePitcherCrop)
	if !ok || bottom.Top || w.setCalls[0].pos.FloorY() != 2 {
		t.Errorf("expected a non-top *DoublePitcherCrop at Y=2, got %#v at Y=%d", w.setCalls[0].blk, w.setCalls[0].pos.FloorY())
	}
	top, ok := w.setCalls[1].blk.(*DoublePitcherCrop)
	if !ok || !top.Top || w.setCalls[1].pos.FloorY() != 3 {
		t.Errorf("expected a top *DoublePitcherCrop at Y=3, got %#v at Y=%d", w.setCalls[1].blk, w.setCalls[1].pos.FloorY())
	}
}

func TestPitcherCropGrowAtMaxAgeFailsWhenSpaceAboveIsBlocked(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	blocker := newTestBlock(false)
	blocker.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = blocker
	p := newTestPitcherCrop(w)
	p.Age = pitcherCropMaxAge

	if p.grow(nil) {
		t.Error("expected grow to fail at max age when the space above is blocked")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
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

func TestPitcherCropOnInteractAtMaxAgeBecomesDoublePitcherCropWithBoneMeal(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	p := newTestPitcherCrop(w)
	p.Age = pitcherCropMaxAge
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if !p.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return true - grow() now handles the max-age DoublePitcherCrop branch")
	}
}
