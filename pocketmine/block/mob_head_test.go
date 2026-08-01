package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

type mobHeadTileWorld struct {
	fakeWorld
	tile *tile.MobHead
}

func (w *mobHeadTileWorld) GetTile(pos Position) (Tile, bool) {
	if w.tile == nil {
		return nil, false
	}
	return w.tile, true
}

func newTestMobHead(w World) *MobHead {
	m := NewMobHead(mustBlockIdentifier(1052), "Test Mob Head", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	m.SetPosition(w, 1, 2, 3)
	return m
}

func TestMobHeadReadStateFromWorldPullsFromTile(t *testing.T) {
	w := &mobHeadTileWorld{}
	headTile := tile.NewMobHead(nil, math.Vector3{})
	headTile.SetMobHeadType(blockutils.MobHeadTypeDragon)
	headTile.SetRotation(9)
	w.tile = headTile

	m := newTestMobHead(w)
	m.ReadStateFromWorld()

	if m.GetMobHeadType() != blockutils.MobHeadTypeDragon {
		t.Errorf("GetMobHeadType() = %v, want Dragon", m.GetMobHeadType())
	}
	if m.GetRotation() != 9 {
		t.Errorf("GetRotation() = %d, want 9", m.GetRotation())
	}
}

func TestMobHeadPlaceRejectsDownFace(t *testing.T) {
	w := &fakeWorld{}
	tx := &fakeBlockTransaction{}
	m := newTestMobHead(w)

	if m.Place(tx, fakeItem{}, m, m, math.Down, math.Vector3{}, nil) {
		t.Error("expected Place to reject a Down face")
	}
}

func TestMobHeadPlaceSetsRotationFromYawOnUpFace(t *testing.T) {
	w := &fakeWorld{}
	tx := &fakeBlockTransaction{}
	m := newTestMobHead(w)

	player := &fakeSignPlayer{} // GetYaw() = 0
	if !m.Place(tx, fakeItem{}, m, m, math.Up, math.Vector3{}, player) {
		t.Fatal("expected Place to succeed on Up")
	}
	if m.Facing != math.Up {
		t.Errorf("Facing = %v, want Up", m.Facing)
	}
	if m.Rotation != 0 {
		t.Errorf("Rotation = %d, want 0 (yaw 0)", m.Rotation)
	}
}

func TestMobHeadSetFacingPanicsOnDown(t *testing.T) {
	w := &fakeWorld{}
	m := newTestMobHead(w)

	defer func() {
		if recover() == nil {
			t.Error("expected SetFacing(Down) to panic")
		}
	}()
	m.SetFacing(math.Down)
}
