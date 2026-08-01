package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestFurnaceSubtypesHaveDistinctSaveIDs(t *testing.T) {
	w := &fakeWorld{}
	if got := NewNormalFurnace(w, math.Vector3{}).SaveID(); got != "Furnace" {
		t.Errorf("NormalFurnace.SaveID() = %q, want %q", got, "Furnace")
	}
	if got := NewBlastFurnace(w, math.Vector3{}).SaveID(); got != "BlastFurnace" {
		t.Errorf("BlastFurnace.SaveID() = %q, want %q", got, "BlastFurnace")
	}
	if got := NewSmoker(w, math.Vector3{}).SaveID(); got != "Smoker" {
		t.Errorf("Smoker.SaveID() = %q, want %q", got, "Smoker")
	}
}

func TestFurnaceSubtypesReportOwnFurnaceType(t *testing.T) {
	w := &fakeWorld{}
	if NewNormalFurnace(w, math.Vector3{}).GetFurnaceType() != FurnaceTypeFurnace {
		t.Error("expected NormalFurnace to report FurnaceTypeFurnace")
	}
	if NewBlastFurnace(w, math.Vector3{}).GetFurnaceType() != FurnaceTypeBlastFurnace {
		t.Error("expected BlastFurnace to report FurnaceTypeBlastFurnace")
	}
	if NewSmoker(w, math.Vector3{}).GetFurnaceType() != FurnaceTypeSmoker {
		t.Error("expected Smoker to report FurnaceTypeSmoker")
	}
}

func TestFurnaceDefaultName(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(w, math.Vector3{})
	if f.GetName() != "Furnace" {
		t.Errorf("GetName() = %q, want %q", f.GetName(), "Furnace")
	}
}

func TestFurnaceSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(w, math.Vector3{})
	f.RemainingFuelTime = 100
	f.CookTime = 50
	f.MaxFuelTime = 200
	f.SetName("Smelter")

	saved := f.SaveNBT()

	decoded := NewFurnace(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.RemainingFuelTime != 100 || decoded.CookTime != 50 || decoded.MaxFuelTime != 200 {
		t.Errorf("got fuel=%d cook=%d max=%d, want 100/50/200", decoded.RemainingFuelTime, decoded.CookTime, decoded.MaxFuelTime)
	}
	if decoded.GetName() != "Smelter" {
		t.Errorf("GetName() = %q, want %q", decoded.GetName(), "Smelter")
	}
}

func TestFurnaceReadSaveDataResetsCookTimeWithoutFuel(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(w, math.Vector3{})

	tag := f.SaveNBT()
	tag.SetShort(FurnaceTagBurnTime, 0)
	tag.SetShort(FurnaceTagCookTime, 30)

	if err := f.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if f.CookTime != 0 {
		t.Errorf("CookTime = %d, want 0 (no fuel remaining should reset it)", f.CookTime)
	}
}

func TestFurnaceReadSaveDataDefaultsMaxFuelTimeToRemaining(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(w, math.Vector3{})

	tag := f.SaveNBT()
	tag.SetShort(FurnaceTagBurnTime, 40)
	tag.RemoveTag(FurnaceTagMaxTime)

	if err := f.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if f.MaxFuelTime != 40 {
		t.Errorf("MaxFuelTime = %d, want 40 (defaults to remaining fuel time)", f.MaxFuelTime)
	}
}

func TestFurnaceCanOpenWithLock(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(w, math.Vector3{})
	f.Lock, f.HasLock = "key", true

	if f.CanOpenWith("wrong") {
		t.Error("expected CanOpenWith to fail with the wrong key")
	}
	if !f.CanOpenWith("key") {
		t.Error("expected CanOpenWith to succeed with the right key")
	}
}
