package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

type signTileWorld struct {
	fakeWorld
	tile *tile.Sign
}

func (w *signTileWorld) GetTile(pos Position) (Tile, bool) {
	if w.tile == nil {
		return nil, false
	}
	return w.tile, true
}

func newTestFloorSign(w World) *FloorSign {
	f := NewFloorSign(mustBlockIdentifier(1047), "Test Floor Sign", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	f.SetPosition(w, 1, 2, 3)
	return f
}

func newTestWallSign(w World) *WallSign {
	ws := NewWallSign(mustBlockIdentifier(1048), "Test Wall Sign", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	ws.SetPosition(w, 1, 2, 3)
	return ws
}

func TestBaseSignReadStateFromWorldPullsStateFromTile(t *testing.T) {
	w := &signTileWorld{}
	signTile := tile.NewSign(nil, math.Vector3{})
	signTile.SetText(blockutils.NewSignText([]string{"hi"}, nil, false))
	signTile.SetWaxed(true)
	signTile.SetEditorEntityRuntimeID(7, true)
	w.tile = signTile

	f := newTestFloorSign(w)
	f.ReadStateFromWorld()

	if lines := f.GetText().GetLines(); lines[0] != "hi" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "hi")
	}
	if !f.IsWaxed() {
		t.Error("expected Waxed to be pulled from the tile")
	}
	id, ok := f.GetEditorEntityRuntimeID()
	if !ok || id != 7 {
		t.Errorf("GetEditorEntityRuntimeID() = (%d, %v), want (7, true)", id, ok)
	}
}

func TestFloorSignPlaceOnlyOnUpFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	f := newTestFloorSign(w)

	if f.Place(tx, fakeItem{}, f, f, math.North, math.Vector3{}, nil) {
		t.Error("expected Place to reject a non-Up face")
	}
	if !f.Place(tx, fakeItem{}, f, f, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on Up")
	}
}

func TestWallSignGetFacingDegrees(t *testing.T) {
	w := &fakeWorld{}
	ws := newTestWallSign(w)

	cases := []struct {
		facing math.Facing
		want   float64
	}{
		{math.South, 0},
		{math.West, 90},
		{math.North, 180},
		{math.East, 270},
	}
	for _, c := range cases {
		ws.Facing = c.facing
		if got := ws.GetFacingDegrees(); got != c.want {
			t.Errorf("facing %v: GetFacingDegrees() = %v, want %v", c.facing, got, c.want)
		}
	}
}

func TestBaseSignWaxPreventsFurtherInteraction(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFloorSign(w)

	honeycomb := fakeItem{typeID: itemTypeIDsHoneycomb}
	fakePlayer := &fakeSignPlayer{}
	if !f.OnInteract(honeycomb, 0, math.Vector3{}, fakePlayer, nil) {
		t.Fatal("expected OnInteract to handle honeycomb")
	}
	if !f.IsWaxed() {
		t.Fatal("expected the sign to become waxed")
	}

	// Once waxed, any further interaction should be a no-op that still reports handled.
	dye := fakeItem{typeID: itemTypeIDsBoneMeal}
	if !f.OnInteract(dye, 0, math.Vector3{}, fakePlayer, nil) {
		t.Error("expected OnInteract to report handled even when waxed")
	}
}

// fakeSignPlayer satisfies the Player interface for BaseSign.OnInteract's frontFace calculation.
type fakeSignPlayer struct{}

func (fakeSignPlayer) ResetFallDistance()                   {}
func (fakeSignPlayer) GetPosition() math.Vector3            { return math.Vector3{X: 1, Y: 2, Z: 4} }
func (fakeSignPlayer) SetOnGround(onGround bool)            {}
func (fakeSignPlayer) GetFallDistance() float64             { return 0 }
func (fakeSignPlayer) SetFallDistance(fallDistance float64) {}
func (fakeSignPlayer) GetBoundingBox() math.AxisAlignedBB   { return math.AxisAlignedBB{} }
func (fakeSignPlayer) GetMotion() math.Vector3              { return math.Vector3{} }
func (fakeSignPlayer) SetOnFire(seconds int)                {}
func (fakeSignPlayer) IsOnFire() bool                       { return false }
func (fakeSignPlayer) Extinguish()                          {}
func (fakeSignPlayer) ExtinguishWithCause(cause int)        {}
func (fakeSignPlayer) CanBeMovedByCurrents() bool           { return true }
func (fakeSignPlayer) Attack(source entity.DamageSource)    {}
func (fakeSignPlayer) GetHorizontalFacing() math.Facing     { return math.North }
func (fakeSignPlayer) IsSneaking() bool                     { return false }
func (fakeSignPlayer) GetYaw() float64                      { return 0 }
func (fakeSignPlayer) GetID() int                           { return 99 }
func (fakeSignPlayer) GetEyePos() math.Vector3              { return math.Vector3{X: 1, Y: 3.62, Z: 4} }
func (fakeSignPlayer) IsSurvival() bool                     { return true }
