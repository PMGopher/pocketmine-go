package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestFloorCoralFan(w World) *FloorCoralFan {
	f := NewFloorCoralFan(mustBlockIdentifier(1035), "Test Floor Coral Fan", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)
	return f
}

func newTestWallCoralFan(w World) *WallCoralFan {
	wcf := NewWallCoralFan(mustBlockIdentifier(1036), "Test Wall Coral Fan", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	wcf.SetPosition(w, 1, 2, 3)
	return wcf
}

func TestFloorCoralFanDeadWithoutNearbyWater(t *testing.T) {
	w := &candleWorld{} // solid, non-water fillers everywhere
	f := newTestFloorCoralFan(w)
	tx := &fakeBlockTransaction{}

	if !f.Place(tx, fakeItem{}, f, f, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed")
	}
	if !f.Dead {
		t.Error("expected the fan to be placed dead without nearby water")
	}
}

func TestFloorCoralFanAliveWithNearbyWater(t *testing.T) {
	w := &neighborWorld{at: [3]int{2, 2, 3}}
	waterBlock := &stemTestBlock{typeID: WATER}
	waterBlock.Block = NewBlock(mustBlockIdentifier(1037), "Water Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	waterBlock.Init(waterBlock)
	waterBlock.SetPosition(w, 2, 2, 3)
	w.neighbor = waterBlock

	f := newTestFloorCoralFan(w)
	tx := &fakeBlockTransaction{}

	if !f.Place(tx, fakeItem{}, f, f, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed")
	}
	if f.Dead {
		t.Error("expected the fan to be placed alive with nearby water")
	}
}

func TestWallCoralFanPlaceRejectsVerticalFace(t *testing.T) {
	w := &candleWorld{}
	wcf := newTestWallCoralFan(w)
	tx := &fakeBlockTransaction{}

	if wcf.Place(tx, fakeItem{}, wcf, wcf, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to reject a vertical face")
	}
}

func TestWallCoralFanPlaceSetsFacingOnHorizontalFace(t *testing.T) {
	w := &candleWorld{}
	wcf := newTestWallCoralFan(w)
	tx := &fakeBlockTransaction{}

	if !wcf.Place(tx, fakeItem{}, wcf, wcf, math.East, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed on a horizontal face with support")
	}
	if wcf.Facing != math.East {
		t.Errorf("Facing = %v, want East", wcf.Facing)
	}
}
