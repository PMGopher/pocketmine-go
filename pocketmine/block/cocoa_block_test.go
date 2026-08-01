package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestCocoaBlock(w World) *CocoaBlock {
	c := NewCocoaBlock(mustBlockIdentifier(1021), "Test Cocoa Block", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func newTestJungleWood(w World) *Wood {
	wd := NewWood(mustBlockIdentifier(1022), "Test Jungle Wood", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeJungle)
	wd.SetPosition(w, 2, 2, 3)
	return wd
}

func TestCocoaBlockPlacesOnlyOnJungleWood(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}

	jungle := newTestJungleWood(w)
	c := newTestCocoaBlock(w)
	if !c.Place(tx, fakeItem{}, jungle, jungle, math.East, math.Vector3{}, nil) {
		t.Error("expected Place to succeed against jungle wood")
	}
	if c.Facing != math.East {
		t.Errorf("Facing = %v, want East", c.Facing)
	}

	oak := NewWood(mustBlockIdentifier(1023), "Test Oak Wood", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), blockutils.WoodTypeOak)
	oak.SetPosition(w, 2, 2, 3)
	c2 := newTestCocoaBlock(w)
	if c2.Place(tx, fakeItem{}, oak, oak, math.East, math.Vector3{}, nil) {
		t.Error("expected Place to fail against non-jungle wood")
	}
}

func TestCocoaBlockPlaceRejectsVerticalFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	jungle := newTestJungleWood(w)

	c := newTestCocoaBlock(w)
	if c.Place(tx, fakeItem{}, jungle, jungle, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to reject a vertical (Up/Down) face")
	}
}

func TestCocoaBlockBreaksWhenLogRemoved(t *testing.T) {
	// The log CocoaBlock attaches to should be behind it (Opposite(Facing)); with only a generic
	// (non-wood) filler there, OnNearbyBlockChange should break the block.
	w := &fakeWorld{}
	c := newTestCocoaBlock(w)
	c.Facing = math.East

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestCocoaBlockStaysWhenLogPresent(t *testing.T) {
	w := &neighborWorld{at: [3]int{0, 2, 3}} // West of (1,2,3) = Opposite(East)
	jungle := newTestJungleWood(w)
	jungle.SetPosition(w, 0, 2, 3)
	w.neighbor = jungle

	c := newTestCocoaBlock(w)
	c.Facing = math.East

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 0 {
		t.Errorf("expected no break while the jungle log is still present, got %d break calls", len(w.breakCalls))
	}
}
