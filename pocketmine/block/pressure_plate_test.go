package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

type fakeLivingEntity struct{}

func (fakeLivingEntity) ResetFallDistance()                   {}
func (fakeLivingEntity) GetPosition() math.Vector3            { return math.Vector3{} }
func (fakeLivingEntity) SetOnGround(onGround bool)            {}
func (fakeLivingEntity) GetFallDistance() float64             { return 0 }
func (fakeLivingEntity) SetFallDistance(fallDistance float64) {}
func (fakeLivingEntity) IsLiving() bool                       { return true }
func (fakeLivingEntity) IsSneaking() bool                     { return false }
func (fakeLivingEntity) GetBoundingBox() math.AxisAlignedBB   { return math.AxisAlignedBB{} }
func (fakeLivingEntity) GetMotion() math.Vector3              { return math.Vector3{} }
func (fakeLivingEntity) SetOnFire(seconds int)                {}
func (fakeLivingEntity) IsOnFire() bool                       { return false }
func (fakeLivingEntity) Extinguish()                          {}
func (fakeLivingEntity) CanBeMovedByCurrents() bool           { return true }
func (fakeLivingEntity) Attack(source entity.DamageSource)    {}

// entityWorld extends fakeWorld with a settable list of nearby entities, for exercising
// PressurePlate.OnScheduledUpdate.
type entityWorld struct {
	fakeWorld
	nearby []Entity
}

func (w *entityWorld) GetNearbyEntities(bb math.AxisAlignedBB) []Entity { return w.nearby }

func newTestStonePressurePlate(w World) *StonePressurePlate {
	idInfo, err := NewBlockIdentifier(1005, nil)
	if err != nil {
		panic(err)
	}
	p := NewStonePressurePlate(idInfo, "Test Stone Pressure Plate", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), 20)
	p.SetPosition(w, 0, 0, 0)
	return p
}

func TestStonePressurePlatePressesForLivingEntityOnly(t *testing.T) {
	w := &entityWorld{}
	plate := newTestStonePressurePlate(w)

	// A non-Living entity shouldn't press the plate (filterIrrelevantEntities excludes it).
	w.nearby = []Entity{fakeItemLikeEntity{}}
	plate.OnScheduledUpdate()
	if w.lastSetBlock != nil {
		t.Fatal("a non-Living entity should not have activated the plate")
	}

	// A Living entity should press it.
	w.nearby = []Entity{fakeLivingEntity{}}
	plate.OnScheduledUpdate()
	pressed, ok := w.lastSetBlock.(*StonePressurePlate)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *StonePressurePlate, got %T", w.lastSetBlock)
	}
	if !pressed.IsPressed() {
		t.Error("plate should be pressed")
	}
	if w.scheduleDelay != plate.DeactivationDelayTicks {
		t.Errorf("scheduled delay = %d, want %d", w.scheduleDelay, plate.DeactivationDelayTicks)
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 sound, got %d", len(w.sounds))
	}
}

// fakeItemLikeEntity satisfies Entity but not Living, to exercise StonePressurePlate's Living-only
// filter.
type fakeItemLikeEntity struct{}

func (fakeItemLikeEntity) ResetFallDistance()                   {}
func (fakeItemLikeEntity) GetPosition() math.Vector3            { return math.Vector3{} }
func (fakeItemLikeEntity) SetOnGround(onGround bool)            {}
func (fakeItemLikeEntity) GetFallDistance() float64             { return 0 }
func (fakeItemLikeEntity) SetFallDistance(fallDistance float64) {}
func (fakeItemLikeEntity) GetBoundingBox() math.AxisAlignedBB   { return math.AxisAlignedBB{} }
func (fakeItemLikeEntity) GetMotion() math.Vector3              { return math.Vector3{} }
func (fakeItemLikeEntity) SetOnFire(seconds int)                {}
func (fakeItemLikeEntity) IsOnFire() bool                       { return false }
func (fakeItemLikeEntity) Extinguish()                          {}
func (fakeItemLikeEntity) CanBeMovedByCurrents() bool           { return true }
func (fakeItemLikeEntity) Attack(source entity.DamageSource)    {}

func TestWeightedPressurePlateSignalStrengthScalesWithEntityCount(t *testing.T) {
	w := &entityWorld{}
	idInfo, err := NewBlockIdentifier(1006, nil)
	if err != nil {
		t.Fatal(err)
	}
	plate := NewWeightedPressurePlate(idInfo, "Test Weighted Pressure Plate", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), 20, 1.0)
	plate.SetPosition(w, 0, 0, 0)

	w.nearby = []Entity{fakeLivingEntity{}, fakeLivingEntity{}, fakeLivingEntity{}}
	plate.OnScheduledUpdate()

	newState, ok := w.lastSetBlock.(*WeightedPressurePlate)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *WeightedPressurePlate, got %T", w.lastSetBlock)
	}
	if newState.GetOutputSignalStrength() != 3 {
		t.Errorf("GetOutputSignalStrength() = %d, want 3", newState.GetOutputSignalStrength())
	}
}
