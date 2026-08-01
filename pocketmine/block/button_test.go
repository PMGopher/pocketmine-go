package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// testButtonType is a stand-in for a concrete leaf type (WoodenButton, StoneButton, etc. — not
// yet ported) that embeds Button. It exists purely to prove the whole embedding chain
// (Block -> Transparent -> Flowable -> Button -> testButtonType) satisfies Behavior end-to-end,
// with self-dispatch and state encoding working the same way a real button would use them.
type testButtonType struct {
	Button
}

func newTestButton(w World) *testButtonType {
	idInfo, err := NewBlockIdentifier(1000, nil)
	if err != nil {
		panic(err)
	}
	t := &testButtonType{Button: Button{
		Flowable:        Flowable{Transparent{NewBlock(idInfo, "Test Button", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}},
		FacingComponent: NewFacingComponent(),
		ActivationTime:  30,
	}}
	t.Init(t)
	t.SetPosition(w, 1, 2, 3)
	return t
}

func (t *testButtonType) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

type fakeWorld struct {
	lastSetPos    Position
	lastSetBlock  Behavior
	schedulePos   math.Vector3
	scheduleDelay int
	sounds        []sound.Sound
	breakCalls    []math.Vector3
}

func (w *fakeWorld) GetBlockAt(x, y, z int) Behavior { return nil }
func (w *fakeWorld) SetBlock(pos Position, blk Behavior) error {
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
}
func (w *fakeWorld) GetTile(pos Position) (Tile, bool)                   { return nil, false }
func (w *fakeWorld) AddTile(tile Tile)                                   {}
func (w *fakeWorld) GetOrLoadChunkAtPosition(pos Position) (Chunk, bool) { return nil, false }
func (w *fakeWorld) AddSound(pos math.Vector3, s sound.Sound)            { w.sounds = append(w.sounds, s) }
func (w *fakeWorld) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {
	w.schedulePos, w.scheduleDelay = pos, delay
}
func (w *fakeWorld) UseBreakOn(pos math.Vector3) bool {
	w.breakCalls = append(w.breakCalls, pos)
	return true
}
func (w *fakeWorld) GetFullLightAt(x, y, z int) int                   { return 15 }
func (w *fakeWorld) GetBlockLightAt(x, y, z int) int                  { return 15 }
func (w *fakeWorld) GetRealBlockSkyLightAt(x, y, z int) int           { return 15 }
func (w *fakeWorld) GetSunAnglePercentage() float64                   { return 0.5 }
func (w *fakeWorld) GetNearbyEntities(bb math.AxisAlignedBB) []Entity { return nil }
func (w *fakeWorld) GetHighestAdjacentFullLightAt(x, y, z int) int    { return 15 }

func TestButtonStateRoundTrips(t *testing.T) {
	w := &fakeWorld{}
	btn := newTestButton(w)
	btn.SetFacing(math.East)
	btn.SetPressed(true)

	data, err := btn.encodeBlockOnlyState()
	if err != nil {
		t.Fatalf("encodeBlockOnlyState: %v", err)
	}

	decoded := newTestButton(w)
	if err := decoded.DecodeBlockOnlyState(data); err != nil {
		t.Fatalf("DecodeBlockOnlyState: %v", err)
	}
	if decoded.GetFacing() != math.East {
		t.Errorf("GetFacing() = %v, want East", decoded.GetFacing())
	}
	if !decoded.IsPressed() {
		t.Error("IsPressed() = false, want true")
	}
}

func TestButtonOnInteractPressesSchedulesAndSounds(t *testing.T) {
	w := &fakeWorld{}
	btn := newTestButton(w)

	if btn.IsPressed() {
		t.Fatal("button should start unpressed")
	}

	btn.OnInteract(nil, math.Up, math.Vector3{}, nil, nil)

	if !btn.IsPressed() {
		t.Fatal("OnInteract should press the button")
	}
	if w.scheduleDelay != btn.ActivationTime {
		t.Errorf("scheduled delay = %d, want %d", w.scheduleDelay, btn.ActivationTime)
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 sound, got %d", len(w.sounds))
	}
	if _, ok := w.sounds[0].(sound.RedstonePowerOnSound); !ok {
		t.Errorf("expected RedstonePowerOnSound, got %T", w.sounds[0])
	}
	if w.lastSetBlock != Behavior(btn) {
		t.Error("SetBlock should have been called with the button itself (via b.self)")
	}

	// Interacting again while already pressed must be a no-op (matches PHP's `if(!$this->pressed)`).
	w.sounds = nil
	btn.OnInteract(nil, math.Up, math.Vector3{}, nil, nil)
	if len(w.sounds) != 0 {
		t.Error("interacting with an already-pressed button should not schedule another sound")
	}

	btn.OnScheduledUpdate()
	if btn.IsPressed() {
		t.Error("OnScheduledUpdate should unpress the button")
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 sound after OnScheduledUpdate, got %d", len(w.sounds))
	}
	if _, ok := w.sounds[0].(sound.RedstonePowerOffSound); !ok {
		t.Errorf("expected RedstonePowerOffSound, got %T", w.sounds[0])
	}
}

func TestButtonCloneIsIndependent(t *testing.T) {
	w := &fakeWorld{}
	original := newTestButton(w)
	original.SetFacing(math.North)

	cloned := original.Clone().(*testButtonType)
	cloned.SetFacing(math.South)
	cloned.SetPressed(true)

	if original.GetFacing() != math.North {
		t.Error("cloning leaked facing changes back into the original")
	}
	if original.IsPressed() {
		t.Error("cloning leaked pressed state back into the original")
	}
}
