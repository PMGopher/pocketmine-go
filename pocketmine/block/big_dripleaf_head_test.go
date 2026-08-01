package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestBigDripleafHead(w World) *BigDripleafHead {
	idInfo, err := NewBlockIdentifier(1011, nil)
	if err != nil {
		panic(err)
	}
	b := NewBigDripleafHead(idInfo, "Test Big Dripleaf Head", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBigDripleafHeadOnScheduledUpdateProgressesThroughTiltStates(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBigDripleafHead(w)

	// Stable: OnScheduledUpdate should be a no-op.
	b.OnScheduledUpdate()
	if b.LeafState != blockutils.DripleafStateStable {
		t.Fatalf("LeafState = %v, want Stable to stay stable", b.LeafState)
	}

	b.LeafState = blockutils.DripleafStateUnstable
	b.OnScheduledUpdate()
	if b.LeafState != blockutils.DripleafStatePartialTilt {
		t.Errorf("LeafState = %v, want PartialTilt after Unstable", b.LeafState)
	}
	if w.scheduleDelay != 10 {
		t.Errorf("scheduleDelay = %d, want 10 (PartialTilt's own delay)", w.scheduleDelay)
	}

	b.OnScheduledUpdate()
	if b.LeafState != blockutils.DripleafStateFullTilt {
		t.Errorf("LeafState = %v, want FullTilt after PartialTilt", b.LeafState)
	}

	b.OnScheduledUpdate()
	if b.LeafState != blockutils.DripleafStateStable {
		t.Errorf("LeafState = %v, want Stable after FullTilt resets", b.LeafState)
	}
}

func TestBigDripleafHeadOnProjectileHitTiltsFully(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBigDripleafHead(w)

	b.OnProjectileHit(nil, math.RayTraceResult{})
	if b.LeafState != blockutils.DripleafStateFullTilt {
		t.Errorf("LeafState = %v, want FullTilt", b.LeafState)
	}

	// Already at FullTilt: a second hit should not re-trigger the tilt-down sound/reschedule.
	w.sounds = nil
	b.OnProjectileHit(nil, math.RayTraceResult{})
	if len(w.sounds) != 0 {
		t.Errorf("expected no additional sound when already at FullTilt, got %d", len(w.sounds))
	}
}

func TestBigDripleafHeadRecalculateCollisionBoxesEmptyWhenFullyTilted(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBigDripleafHead(w)
	b.LeafState = blockutils.DripleafStateFullTilt

	if boxes := b.RecalculateCollisionBoxes(); boxes != nil {
		t.Errorf("expected no collision boxes at FullTilt, got %v", boxes)
	}
}
