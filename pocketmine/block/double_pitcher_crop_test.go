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
