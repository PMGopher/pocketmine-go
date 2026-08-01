package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestOminousFloorBanner(w World) *OminousFloorBanner {
	o := NewOminousFloorBanner(mustBlockIdentifier(1053), "Test Ominous Floor Banner", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	o.SetPosition(w, 1, 2, 3)
	return o
}

func newTestOminousWallBanner(w World) *OminousWallBanner {
	o := NewOminousWallBanner(mustBlockIdentifier(1054), "Test Ominous Wall Banner", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	o.SetPosition(w, 1, 2, 3)
	return o
}

func TestOminousFloorBannerPlaceOnlyOnUpFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	o := newTestOminousFloorBanner(w)

	if o.Place(tx, fakeItem{}, o, o, math.East, math.Vector3{}, nil) {
		t.Error("expected Place to reject a non-Up face")
	}
	if !o.Place(tx, fakeItem{}, o, o, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on Up with solid support below")
	}
}

func TestOminousFloorBannerGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	o := newTestOminousFloorBanner(w)
	if o.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", o.GetFuelTime())
	}
}

func TestOminousWallBannerPlaceRejectsVerticalFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	o := newTestOminousWallBanner(w)

	if o.Place(tx, fakeItem{}, o, o, math.Down, math.Vector3{}, nil) {
		t.Error("expected Place to reject a vertical face")
	}
	if !o.Place(tx, fakeItem{}, o, o, math.West, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on a horizontal face with solid support")
	}
	if o.Facing != math.West {
		t.Errorf("Facing = %v, want West", o.Facing)
	}
}

func TestOminousBannerOnNearbyBlockChangeBreaksWithoutSolidSupport(t *testing.T) {
	w := &noSupportWorld{}
	o := newTestOminousFloorBanner(w)

	o.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}
