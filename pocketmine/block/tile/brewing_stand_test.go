package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestBrewingStandDefaultName(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrewingStand(w, math.Vector3{})
	if b.GetName() != "Brewing Stand" {
		t.Errorf("GetName() = %q, want %q", b.GetName(), "Brewing Stand")
	}
}

func TestBrewingStandSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrewingStand(w, math.Vector3{})
	b.BrewTime = 123
	b.MaxFuelTime = 20
	b.RemainingFuelTime = 15

	decoded := NewBrewingStand(w, math.Vector3{})
	if err := decoded.ReadSaveData(b.SaveNBT()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.BrewTime != 123 || decoded.MaxFuelTime != 20 || decoded.RemainingFuelTime != 15 {
		t.Errorf("got brew=%d max=%d remaining=%d, want 123/20/15", decoded.BrewTime, decoded.MaxFuelTime, decoded.RemainingFuelTime)
	}
}

func TestBrewingStandReadSaveDataPrefersLegacyBrewTimeTag(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrewingStand(w, math.Vector3{})
	b.RemainingFuelTime = 5 // keep fuel nonzero so the zero-fuel reset doesn't clear BrewTime

	tag := b.SaveNBT()
	tag.SetShort(BrewingStandTagBrewTime, 1)        // "CookTime" (PE)
	tag.SetShort(BrewingStandTagBrewTimeLegacy, 99) // "BrewTime" (legacy) takes priority

	if err := b.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.BrewTime != 99 {
		t.Errorf("BrewTime = %d, want 99 (legacy tag should take priority)", b.BrewTime)
	}
}

func TestBrewingStandReadSaveDataZeroesEverythingWithoutFuel(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrewingStand(w, math.Vector3{})

	tag := b.SaveNBT()
	tag.SetShort(BrewingStandTagBrewTime, 50)
	tag.SetShort(BrewingStandTagMaxFuelTime, 20)
	tag.RemoveTag(BrewingStandTagRemainingFuelTime, BrewingStandTagRemainingFuelTimeLegacy)

	if err := b.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.BrewTime != 0 || b.MaxFuelTime != 0 || b.RemainingFuelTime != 0 {
		t.Errorf("got brew=%d max=%d remaining=%d, want all 0 with no fuel remaining", b.BrewTime, b.MaxFuelTime, b.RemainingFuelTime)
	}
}

func TestBrewingStandReadSaveDataDefaultsMaxFuelTimeToRemaining(t *testing.T) {
	w := &fakeWorld{}
	b := NewBrewingStand(w, math.Vector3{})

	tag := b.SaveNBT()
	tag.SetByte(BrewingStandTagRemainingFuelTimeLegacy, 12)
	tag.RemoveTag(BrewingStandTagMaxFuelTime)

	if err := b.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.RemainingFuelTime != 12 {
		t.Errorf("RemainingFuelTime = %d, want 12", b.RemainingFuelTime)
	}
	if b.MaxFuelTime != 12 {
		t.Errorf("MaxFuelTime = %d, want 12 (defaults to remaining fuel time)", b.MaxFuelTime)
	}
}
