package entity

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestHuman() *Human {
	return NewHuman(math.NewVector3(1, 2, 3), math.OneAABB(), "test-uuid")
}

func TestNewHumanStoresUUID(t *testing.T) {
	h := newTestHuman()
	if got := h.GetUniqueID(); got != "test-uuid" {
		t.Errorf("GetUniqueID() = %q, want %q", got, "test-uuid")
	}
}

func TestHumanGetEyePosOffsetsByEyeHeight(t *testing.T) {
	h := newTestHuman()
	want := math.NewVector3(1, 2+humanEyeHeight, 3)
	if got := h.GetEyePos(); got != want {
		t.Errorf("GetEyePos() = %v, want %v", got, want)
	}
}

func TestHumanGetEyeHeightMatchesTheRealVanillaValue(t *testing.T) {
	h := newTestHuman()
	if got := h.GetEyeHeight(); got != 1.62 {
		t.Errorf("GetEyeHeight() = %v, want 1.62", got)
	}
}

func TestHumanIsLivingAndInheritsEntityBehaviour(t *testing.T) {
	h := newTestHuman()
	if !h.IsLiving() {
		t.Error("IsLiving() = false, want true")
	}
	h.SetSneaking(true)
	if !h.IsSneaking() {
		t.Error("IsSneaking() = false after SetSneaking(true)")
	}
	h.SetFallDistance(3)
	if h.GetFallDistance() != 3 {
		t.Errorf("GetFallDistance() = %v, want 3 (inherited from Entity)", h.GetFallDistance())
	}
}
