package block

import (
	"testing"
)

func newTestDoublePitcherCrop(w World) *DoublePitcherCrop {
	d := NewDoublePitcherCrop(mustBlockIdentifier(1064), "Test Double Pitcher Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	d.SetPosition(w, 1, 2, 3)
	return d
}

func TestDoublePitcherCropTicksRandomlyOnlyOnBottomAndBelowMaxAge(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDoublePitcherCrop(w)

	if !d.TicksRandomly() {
		t.Error("expected a fresh bottom-half, age-0 DoublePitcherCrop to tick randomly")
	}

	d.SetTop(true)
	if d.TicksRandomly() {
		t.Error("expected the top half not to tick randomly")
	}

	d.SetTop(false)
	d.SetAge(doublePitcherCropMaxAge)
	if d.TicksRandomly() {
		t.Error("expected a fully-grown bottom half not to tick randomly")
	}
}

func TestDoublePitcherCropCollisionBoxesEmptyOnTop(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDoublePitcherCrop(w)

	if boxes := d.RecalculateCollisionBoxes(); len(boxes) != 1 {
		t.Errorf("expected exactly one collision box on the bottom half, got %d", len(boxes))
	}

	d.SetTop(true)
	if boxes := d.RecalculateCollisionBoxes(); boxes != nil {
		t.Errorf("expected no collision box on the top half, got %v", boxes)
	}
}

func TestDoublePitcherCropSetAgeRejectsOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	d := newTestDoublePitcherCrop(w)

	defer func() {
		if recover() == nil {
			t.Error("expected SetAge to panic for an out-of-range value")
		}
	}()
	d.SetAge(doublePitcherCropMaxAge + 1)
}

func TestDoublePitcherCropGrowFromBottomRebuildsBothHalves(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	// The top half (Y+1) must be AIR or the same type - default AIR filler satisfies this.
	d := newTestDoublePitcherCrop(w)
	d.Age = 0

	if !d.grow(nil) {
		t.Fatal("expected grow to succeed")
	}
	if len(w.setCalls) != 2 {
		t.Fatalf("expected 2 SetBlock calls, got %d", len(w.setCalls))
	}
	bottom, ok := w.setCalls[0].blk.(*DoublePitcherCrop)
	if !ok || bottom.Top || bottom.Age != 1 || w.setCalls[0].pos.FloorY() != 2 {
		t.Errorf("expected a non-top *DoublePitcherCrop age 1 at Y=2, got %#v at Y=%d", w.setCalls[0].blk, w.setCalls[0].pos.FloorY())
	}
	top, ok := w.setCalls[1].blk.(*DoublePitcherCrop)
	if !ok || !top.Top || top.Age != 1 || w.setCalls[1].pos.FloorY() != 3 {
		t.Errorf("expected a top *DoublePitcherCrop age 1 at Y=3, got %#v at Y=%d", w.setCalls[1].blk, w.setCalls[1].pos.FloorY())
	}
}

func TestDoublePitcherCropGrowFromTopRebuildsBothHalves(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	d := NewDoublePitcherCrop(mustBlockIdentifier(1064), "Test Double Pitcher Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	d.SetPosition(w, 1, 3, 3) // this is the top half; bottom is Y=2
	d.SetTop(true)
	d.Age = 0

	if !d.grow(nil) {
		t.Fatal("expected grow to succeed")
	}
	if len(w.setCalls) != 2 {
		t.Fatalf("expected 2 SetBlock calls, got %d", len(w.setCalls))
	}
	if w.setCalls[0].pos.FloorY() != 2 || w.setCalls[1].pos.FloorY() != 3 {
		t.Errorf("expected bottom at Y=2 and top at Y=3, got Y=%d then Y=%d", w.setCalls[0].pos.FloorY(), w.setCalls[1].pos.FloorY())
	}
}

func TestDoublePitcherCropGrowReturnsFalseAtMaxAge(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	d := newTestDoublePitcherCrop(w)
	d.Age = doublePitcherCropMaxAge

	if d.grow(nil) {
		t.Error("expected grow to fail at max age")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestDoublePitcherCropGrowFailsWhenOtherHalfBlocked(t *testing.T) {
	w := &pitcherCropGrowWorld{blocks: map[[3]int]Behavior{}}
	blocker := newTestBlock(false)
	blocker.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = blocker
	d := newTestDoublePitcherCrop(w)
	d.Age = 0

	if d.grow(nil) {
		t.Error("expected grow to fail when the other half of the plant is blocked")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}
