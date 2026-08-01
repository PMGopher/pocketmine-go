package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestHopper(w World) *Hopper {
	h := NewHopper(mustBlockIdentifier(1070), "Test Hopper", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	h.SetPosition(w, 1, 2, 3)
	return h
}

func TestHopperDefaultFacingIsDown(t *testing.T) {
	w := &fakeWorld{}
	h := newTestHopper(w)
	if h.GetFacing() != math.Down {
		t.Errorf("GetFacing() = %v, want Down", h.GetFacing())
	}
}

func TestHopperSetFacingRejectsUp(t *testing.T) {
	w := &fakeWorld{}
	h := newTestHopper(w)
	defer func() {
		if recover() == nil {
			t.Error("expected SetFacing(Up) to panic")
		}
	}()
	h.SetFacing(math.Up)
}

func TestHopperPlaceFacesOppositeClickedFaceUnlessDown(t *testing.T) {
	w := &fakeWorld{}
	tx := &fakeBlockTransaction{}

	h1 := newTestHopper(w)
	h1.Place(tx, fakeItem{}, h1, h1, math.Down, math.Vector3{}, nil)
	if h1.Facing != math.Down {
		t.Errorf("Facing = %v, want Down (clicked face was Down)", h1.Facing)
	}

	h2 := newTestHopper(w)
	h2.Place(tx, fakeItem{}, h2, h2, math.North, math.Vector3{}, nil)
	if h2.Facing != math.South {
		t.Errorf("Facing = %v, want South (opposite of clicked North face)", h2.Facing)
	}
}

func TestHopperGetSupportType(t *testing.T) {
	w := &fakeWorld{}
	h := newTestHopper(w) // Facing defaults to Down

	if got := h.GetSupportType(math.Up); got != blockutils.SupportTypeFull {
		t.Errorf("GetSupportType(Up) = %v, want Full", got)
	}
	if got := h.GetSupportType(math.Down); got != blockutils.SupportTypeCenter {
		t.Errorf("GetSupportType(Down) with Facing=Down = %v, want Center", got)
	}

	h.Facing = math.North
	if got := h.GetSupportType(math.Down); got != blockutils.SupportTypeNone {
		t.Errorf("GetSupportType(Down) with Facing=North = %v, want None", got)
	}
}

func TestHopperOnInteractReturnsFalseWithoutPlayer(t *testing.T) {
	w := &fakeWorld{}
	h := newTestHopper(w)
	if h.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false when there's no player")
	}
}

func TestHopperOnInteractReturnsTrueWithPlayer(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	h := newTestHopper(w)
	w.tiles[[3]int{1, 2, 3}] = tile.NewHopper(w, math.NewVector3(1, 2, 3))

	if !h.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true when a player is present")
	}
}
