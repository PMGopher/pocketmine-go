package block

import (
	"testing"
)

func newTestFire(w World) *Fire {
	f := NewFire(mustBlockIdentifier(1031), "Test Fire", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)
	return f
}

func newTestSoulFire(w World) *SoulFire {
	s := NewSoulFire(mustBlockIdentifier(1032), "Test Soul Fire", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

type fireTrackingEntity struct {
	fakeItemLikeEntity
	onFireSeconds int
}

func (f *fireTrackingEntity) SetOnFire(seconds int) { f.onFireSeconds = seconds }

func TestBaseFireOnEntityInsideSetsFireDuration(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFire(w)

	e := &fireTrackingEntity{}
	if !f.OnEntityInside(e) {
		t.Fatal("expected OnEntityInside to report handled")
	}
	if e.onFireSeconds != 8 {
		t.Errorf("onFireSeconds = %d, want 8", e.onFireSeconds)
	}
}

func TestSoulFireGetLightLevelAndFireDamage(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSoulFire(w)

	if s.GetLightLevel() != 10 {
		t.Errorf("GetLightLevel() = %d, want 10", s.GetLightLevel())
	}
	if s.GetFireDamage() != 2 {
		t.Errorf("GetFireDamage() = %d, want 2", s.GetFireDamage())
	}
}

func TestSoulFireBreaksWithoutSoulSandOrSoulSoil(t *testing.T) {
	// candleWorld's GetBlockAt returns a solid (non-soul-sand/soil) filler rather than nil, which
	// SoulFire.OnNearbyBlockChange needs since it dereferences the block below directly.
	w := &candleWorld{}
	s := newTestSoulFire(w)

	s.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestFireGetFireDamage(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFire(w)
	if f.GetFireDamage() != 1 {
		t.Errorf("GetFireDamage() = %d, want 1", f.GetFireDamage())
	}
}

func TestFireHasAdjacentFlammableBlocksFalseWithPlainNeighbors(t *testing.T) {
	w := &candleWorld{}
	f := newTestFire(w)

	if f.hasAdjacentFlammableBlocks() {
		t.Error("expected no flammable neighbours (candleWorld's filler blocks have zero flammability)")
	}
}
