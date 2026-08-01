package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestBell(w World) *Bell {
	idInfo, err := NewBlockIdentifier(1007, nil)
	if err != nil {
		panic(err)
	}
	b := NewBell(idInfo, "Test Bell", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBellPlaceOnFloorAndCeiling(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}

	floor := newTestBell(w)
	if !floor.Place(tx, fakeItem{}, floor, floor, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected floor placement to succeed")
	}
	if floor.AttachmentType != blockutils.BellAttachmentTypeFloor {
		t.Errorf("AttachmentType = %v, want Floor", floor.AttachmentType)
	}

	ceiling := newTestBell(w)
	if !ceiling.Place(tx, fakeItem{}, ceiling, ceiling, math.Down, math.Vector3{}, nil) {
		t.Fatal("expected ceiling placement to succeed")
	}
	if ceiling.AttachmentType != blockutils.BellAttachmentTypeCeiling {
		t.Errorf("AttachmentType = %v, want Ceiling", ceiling.AttachmentType)
	}
}

func TestBellFloorOnlyRingsOnMatchingAxis(t *testing.T) {
	w := &candleWorld{}
	b := newTestBell(w)
	b.AttachmentType = blockutils.BellAttachmentTypeFloor
	b.Facing = math.North // axis Z

	if !b.isValidFaceToRing(math.South) {
		t.Error("expected a hit sharing the bell's facing axis (Z) to be valid")
	}
	if b.isValidFaceToRing(math.East) {
		t.Error("expected a hit on a different axis (X) to be invalid")
	}
}

func TestBellWallRingsOnPerpendicularFaces(t *testing.T) {
	w := &candleWorld{}
	b := newTestBell(w)
	b.AttachmentType = blockutils.BellAttachmentTypeTwoWalls
	b.Facing = math.North

	if !b.isValidFaceToRing(math.East) || !b.isValidFaceToRing(math.West) {
		t.Error("expected the two side faces perpendicular to facing to be valid")
	}
	if b.isValidFaceToRing(math.North) || b.isValidFaceToRing(math.South) {
		t.Error("expected the facing axis itself not to be valid for a wall-mounted bell")
	}
}

func TestBellCeilingAlwaysRings(t *testing.T) {
	w := &candleWorld{}
	b := newTestBell(w)
	b.AttachmentType = blockutils.BellAttachmentTypeCeiling

	for _, f := range math.AllFacing {
		if !b.isValidFaceToRing(f) {
			t.Errorf("expected ceiling bell to accept every face, but rejected %v", f)
		}
	}
}
