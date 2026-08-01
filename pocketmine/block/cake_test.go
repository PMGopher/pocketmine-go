package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// cakeAirWorld returns an Air filler for every GetBlockAt call, so Cake's canBeSupportedAt (which
// checks GetTypeId() != AIR, not GetSupportType) sees no support.
type cakeAirWorld struct {
	fakeWorld
	breakCalls []math.Vector3
}

func (w *cakeAirWorld) GetBlockAt(x, y, z int) Behavior {
	air := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	air.SetPosition(w, x, y, z)
	return air
}

func (w *cakeAirWorld) UseBreakOn(pos math.Vector3) bool {
	w.breakCalls = append(w.breakCalls, pos)
	return true
}

func newTestCake(w World) *Cake {
	c := NewCake(mustBlockIdentifier(1056), "Test Cake", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCakeCanBePlacedAtRequiresNonAirSupport(t *testing.T) {
	solid := &candleWorld{}
	c := newTestCake(solid)
	replace := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace.SetPosition(solid, 1, 2, 3)
	if !c.CanBePlacedAt(replace, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to accept a non-air block below")
	}

	air := &cakeAirWorld{}
	c2 := newTestCake(air)
	replace2 := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace2.SetPosition(air, 1, 2, 3)
	if c2.CanBePlacedAt(replace2, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to reject air below")
	}
}

func TestCakeOnNearbyBlockChangeBreaksWithoutSupport(t *testing.T) {
	w := &cakeAirWorld{}
	c := newTestCake(w)

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestCakeSetBitesRejectsOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCake(w)

	defer func() {
		if recover() == nil {
			t.Error("expected SetBites to panic for an out-of-range value")
		}
	}()
	c.SetBites(cakeMaxBites + 1)
}

func TestCakeSetBitesAccepted(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCake(w)
	c.SetBites(3)
	if c.GetBites() != 3 {
		t.Errorf("GetBites() = %d, want 3", c.GetBites())
	}
}

func TestCakeWithCandleOnInteractLightsWithFlintAndSteel(t *testing.T) {
	w := &candleWorld{}
	c := NewCakeWithCandle(mustBlockIdentifier(1057), "Test Cake With Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	flintAndSteel := fakeItem{typeID: itemTypeIDsFlintAndSteel}
	if !c.OnInteract(flintAndSteel, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle flint and steel")
	}
	if !c.IsLit() {
		t.Error("flint and steel should have lit the candle")
	}
}

func TestCakeWithCandleBlocksNonUpInteractWhileLit(t *testing.T) {
	w := &fakeWorld{}
	c := NewCakeWithCandle(mustBlockIdentifier(1057), "Test Cake With Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	c.SetLit(true)

	if !c.OnInteract(fakeItem{}, math.East, math.Vector3{}, nil, nil) {
		t.Error("expected a lit candle to swallow non-Up interactions")
	}
}

func TestCakeWithCandleGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	c := NewCakeWithCandle(mustBlockIdentifier(1057), "Test Cake With Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	if c.GetLightLevel() != 0 {
		t.Errorf("GetLightLevel() = %d, want 0 while unlit", c.GetLightLevel())
	}
	c.SetLit(true)
	if c.GetLightLevel() != 3 {
		t.Errorf("GetLightLevel() = %d, want 3 while lit", c.GetLightLevel())
	}
}

func TestCakeWithDyedCandleDefaultsToWhite(t *testing.T) {
	w := &fakeWorld{}
	c := NewCakeWithDyedCandle(mustBlockIdentifier(1058), "Test Cake With Dyed Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	if c.GetColor() != blockutils.DyeColorWhite {
		t.Errorf("GetColor() = %v, want White", c.GetColor())
	}
}
