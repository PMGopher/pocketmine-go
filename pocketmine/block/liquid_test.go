package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
)

func newTestWater(w World) *Water {
	wa := NewWater(mustBlockIdentifier(1041), "Test Water", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	wa.SetPosition(w, 1, 2, 3)
	return wa
}

func newTestLava(w World) *Lava {
	l := NewLava(mustBlockIdentifier(1042), "Test Lava", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	l.SetPosition(w, 1, 2, 3)
	return l
}

func TestLiquidIsSource(t *testing.T) {
	w := &fakeWorld{}
	wa := newTestWater(w)

	if !wa.IsSource() {
		t.Error("a fresh liquid (Decay=0, Falling=false) should be a source")
	}
	wa.Decay = 3
	if wa.IsSource() {
		t.Error("a liquid with nonzero decay should not be a source")
	}
	wa.Decay = 0
	wa.Falling = true
	if wa.IsSource() {
		t.Error("a falling liquid should not be a source")
	}
}

func TestLiquidGetFluidHeightPercent(t *testing.T) {
	w := &fakeWorld{}
	wa := newTestWater(w)

	wa.Decay = 0
	if got, want := wa.GetFluidHeightPercent(), 1.0/9; got != want {
		t.Errorf("source height = %v, want %v", got, want)
	}
	wa.Decay = 7
	if got, want := wa.GetFluidHeightPercent(), 8.0/9; got != want {
		t.Errorf("decay=7 height = %v, want %v", got, want)
	}
	wa.Falling = true
	if got, want := wa.GetFluidHeightPercent(), 1.0/9; got != want {
		t.Errorf("falling height = %v, want %v (decay ignored while falling)", got, want)
	}
}

func TestLiquidGetStillFormAndFlowingForm(t *testing.T) {
	w := &fakeWorld{}
	wa := newTestWater(w)
	wa.Still = false

	still := wa.GetStillForm()
	stillWater, ok := still.(*Water)
	if !ok || !stillWater.Still {
		t.Errorf("GetStillForm() = %#v, want a *Water with Still=true", still)
	}
	// The original should be untouched.
	if wa.Still {
		t.Error("GetStillForm should not mutate the original block")
	}

	wa.Still = true
	flowing := wa.GetFlowingForm()
	flowingWater, ok := flowing.(*Water)
	if !ok || flowingWater.Still {
		t.Errorf("GetFlowingForm() = %#v, want a *Water with Still=false", flowing)
	}
}

type onFireTrackingEntity struct {
	fakeItemLikeEntity
	onFireSeconds   int
	extinguished    bool
	extinguishCause int
	resetCalled     bool
	reportsOnFire   bool
}

func (e *onFireTrackingEntity) SetOnFire(seconds int) { e.onFireSeconds = seconds }
func (e *onFireTrackingEntity) Extinguish()           { e.extinguished = true }
func (e *onFireTrackingEntity) ExtinguishWithCause(cause int) {
	e.extinguished = true
	e.extinguishCause = cause
}
func (e *onFireTrackingEntity) IsOnFire() bool     { return e.reportsOnFire }
func (e *onFireTrackingEntity) ResetFallDistance() { e.resetCalled = true }

func TestWaterOnEntityInsideExtinguishesBurningEntity(t *testing.T) {
	w := &fakeWorld{}
	wa := newTestWater(w)

	e := &onFireTrackingEntity{reportsOnFire: true}
	if !wa.OnEntityInside(e) {
		t.Fatal("expected OnEntityInside to report handled")
	}
	if !e.resetCalled {
		t.Error("expected fall distance to be reset")
	}
	if !e.extinguished {
		t.Error("expected a burning entity to be extinguished")
	}
	if e.extinguishCause != entity.EntityExtinguishCauseWater {
		t.Errorf("extinguishCause = %d, want EntityExtinguishCauseWater", e.extinguishCause)
	}
}

func TestWaterOnEntityInsideLeavesNonBurningEntityAlone(t *testing.T) {
	w := &fakeWorld{}
	wa := newTestWater(w)

	e := &onFireTrackingEntity{reportsOnFire: false}
	wa.OnEntityInside(e)
	if e.extinguished {
		t.Error("expected a non-burning entity not to be extinguished")
	}
}

func TestLavaOnEntityInsideIgnites(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLava(w)

	e := &onFireTrackingEntity{}
	if !l.OnEntityInside(e) {
		t.Fatal("expected OnEntityInside to report handled")
	}
	if e.onFireSeconds != 8 {
		t.Errorf("onFireSeconds = %d, want 8", e.onFireSeconds)
	}
	if !e.resetCalled {
		t.Error("expected fall distance to be reset")
	}
}
