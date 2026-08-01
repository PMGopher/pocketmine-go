package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// farmlandWorld returns a filler with GetTypeId()==FARMLAND for every GetBlockAt call.
type farmlandWorld struct {
	fakeWorld
	breakCalls []math.Vector3
}

func (w *farmlandWorld) GetBlockAt(x, y, z int) Behavior {
	filler := newTestBlock(false)
	filler.idInfo = mustBlockIdentifier(FARMLAND)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *farmlandWorld) UseBreakOn(pos math.Vector3) bool {
	w.breakCalls = append(w.breakCalls, pos)
	return true
}

func newTestTorchflowerCrop(w World) *TorchflowerCrop {
	t := NewTorchflowerCrop(mustBlockIdentifier(1062), "Test Torchflower Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	t.SetPosition(w, 1, 2, 3)
	return t
}

func TestTorchflowerCropCanBePlacedAtRequiresFarmland(t *testing.T) {
	farmland := &farmlandWorld{}
	tc := newTestTorchflowerCrop(farmland)
	replace := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace.SetPosition(farmland, 1, 2, 3)
	if !tc.CanBePlacedAt(replace, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to accept farmland below")
	}

	notFarmland := &candleWorld{}
	tc2 := newTestTorchflowerCrop(notFarmland)
	replace2 := NewAir(mustBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	replace2.SetPosition(notFarmland, 1, 2, 3)
	if tc2.CanBePlacedAt(replace2, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to reject non-farmland below")
	}
}

func TestTorchflowerCropOnNearbyBlockChangeBreaksWithoutFarmland(t *testing.T) {
	w := &cakeAirWorld{}
	tc := newTestTorchflowerCrop(w)

	tc.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestTorchflowerCropReadyState(t *testing.T) {
	w := &fakeWorld{}
	tc := newTestTorchflowerCrop(w)

	if tc.IsReady() {
		t.Error("expected a fresh TorchflowerCrop not to be ready")
	}
	tc.SetReady(true)
	if !tc.IsReady() {
		t.Error("expected SetReady(true) to make IsReady() true")
	}
}
