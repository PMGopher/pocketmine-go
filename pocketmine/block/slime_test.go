package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

type fakeMotionEntity struct {
	fakeItemLikeEntity
	motion       math.Vector3
	resetCalls   int
	sneaking     bool
	livingMarker bool
}

func (f *fakeMotionEntity) GetMotion() math.Vector3 { return f.motion }
func (f *fakeMotionEntity) ResetFallDistance()      { f.resetCalls++ }
func (f *fakeMotionEntity) IsLiving() bool          { return f.livingMarker }
func (f *fakeMotionEntity) IsSneaking() bool        { return f.sneaking }

func newTestSlime(w World) *Slime {
	idInfo, err := NewBlockIdentifier(1015, nil)
	if err != nil {
		panic(err)
	}
	s := NewSlime(idInfo, "Test Slime", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestSlimeBouncesNonSneakingEntity(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSlime(w)

	e := &fakeMotionEntity{motion: math.Vector3{Y: -0.5}, livingMarker: true, sneaking: false}
	bounce, handled := s.OnEntityLand(e)
	if !handled {
		t.Fatal("expected OnEntityLand to report handled")
	}
	if bounce != 0.5 {
		t.Errorf("bounce = %v, want 0.5 (negated downward motion)", bounce)
	}
	if e.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", e.resetCalls)
	}
}

func TestSlimeDoesNotBounceSneakingLivingEntity(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSlime(w)

	e := &fakeMotionEntity{motion: math.Vector3{Y: -0.5}, livingMarker: true, sneaking: true}
	_, handled := s.OnEntityLand(e)
	if handled {
		t.Error("expected a sneaking Living entity not to bounce")
	}
	if e.resetCalls != 0 {
		t.Errorf("resetCalls = %d, want 0 (should not reset fall distance when not bouncing)", e.resetCalls)
	}
}
