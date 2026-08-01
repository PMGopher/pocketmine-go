package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestHopperDefaultName(t *testing.T) {
	w := &fakeWorld{}
	h := NewHopper(w, math.Vector3{})
	if h.GetName() != "Hopper" {
		t.Errorf("GetName() = %q, want %q", h.GetName(), "Hopper")
	}
}

func TestHopperSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	h := NewHopper(w, math.Vector3{})
	h.TransferCooldown = 8
	h.SetName("Feeder")

	decoded := NewHopper(w, math.Vector3{})
	if err := decoded.ReadSaveData(h.SaveNBT()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.TransferCooldown != 8 {
		t.Errorf("TransferCooldown = %d, want 8", decoded.TransferCooldown)
	}
	if decoded.GetName() != "Feeder" {
		t.Errorf("GetName() = %q, want %q", decoded.GetName(), "Feeder")
	}
}

func TestHopperCanOpenWithLock(t *testing.T) {
	w := &fakeWorld{}
	h := NewHopper(w, math.Vector3{})
	h.Lock, h.HasLock = "key", true

	if h.CanOpenWith("wrong") {
		t.Error("expected CanOpenWith to fail with the wrong key")
	}
	if !h.CanOpenWith("key") {
		t.Error("expected CanOpenWith to succeed with the right key")
	}
}
