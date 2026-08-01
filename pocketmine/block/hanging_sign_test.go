package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestCeilingCenterHangingSign(w World) *CeilingCenterHangingSign {
	c := NewCeilingCenterHangingSign(mustBlockIdentifier(1049), "Test Ceiling Center Hanging Sign", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	c.SetPosition(w, 1, 2, 3)
	return c
}

func newTestCeilingEdgesHangingSign(w World) *CeilingEdgesHangingSign {
	c := NewCeilingEdgesHangingSign(mustBlockIdentifier(1050), "Test Ceiling Edges Hanging Sign", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	c.SetPosition(w, 1, 2, 3)
	return c
}

func newTestWallHangingSign(w World) *WallHangingSign {
	wh := NewWallHangingSign(mustBlockIdentifier(1051), "Test Wall Hanging Sign", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	wh.SetPosition(w, 1, 2, 3)
	return wh
}

func TestCeilingCenterHangingSignPlaceOnlyOnDownFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	c := newTestCeilingCenterHangingSign(w)

	if c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to reject a non-Down face")
	}
	if !c.Place(tx, fakeItem{}, c, c, math.Down, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on Down with center support above")
	}
}

func TestCeilingCenterHangingSignBreaksWithoutSupport(t *testing.T) {
	w := &noSupportWorld{}
	c := newTestCeilingCenterHangingSign(w)

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn once, got %d", len(w.breakCalls))
	}
}

func TestCeilingEdgesHangingSignPlaceSetsFacingOpposite(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	c := newTestCeilingEdgesHangingSign(w)

	player := &fakeSignPlayer{}
	if !c.Place(tx, fakeItem{}, c, c, math.Down, math.Vector3{}, player) {
		t.Fatal("expected Place to succeed on Down with full support above")
	}
	if c.Facing != math.Opposite(math.North) { // fakeSignPlayer.GetHorizontalFacing() = North
		t.Errorf("Facing = %v, want %v", c.Facing, math.Opposite(math.North))
	}
}

func TestWallHangingSignOnNearbyBlockChangeIsNoop(t *testing.T) {
	w := &noSupportWorld{}
	wh := newTestWallHangingSign(w)

	wh.OnNearbyBlockChange()

	if len(w.breakCalls) != 0 {
		t.Errorf("expected OnNearbyBlockChange to never break (disabled self-destruct), got %d break calls", len(w.breakCalls))
	}
}

func TestWallHangingSignGetSupportingFace(t *testing.T) {
	w := &fakeWorld{}
	wh := newTestWallHangingSign(w)
	wh.Facing = math.North

	if got, want := wh.GetSupportingFace(), math.RotateY(math.North, true); got != want {
		t.Errorf("GetSupportingFace() = %v, want %v", got, want)
	}
}

func TestWallHangingSignPlaceRequiresPlayer(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	wh := newTestWallHangingSign(w)

	if wh.Place(tx, fakeItem{}, wh, wh, math.North, math.Vector3{}, nil) {
		t.Error("expected Place to fail without a player")
	}
}
