package item

import "testing"

var (
	_ Item = (*GoatHorn)(nil)
	_ Item = (*SuspiciousStew)(nil)
	_ Item = (*Medicine)(nil)
)

func TestGoatHornDefaultsAndState(t *testing.T) {
	g := NewGoatHorn(NewItemIdentifier(GOAT_HORN), "Goat Horn")
	if g.GetHornType() != GoatHornTypePonder {
		t.Errorf("GetHornType() = %v, want Ponder", g.GetHornType())
	}
	if g.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", g.GetMaxStackSize())
	}
	if g.GetCooldownTicks() != 140 {
		t.Errorf("GetCooldownTicks() = %d, want 140", g.GetCooldownTicks())
	}
	tag, ok := g.GetCooldownTag()
	if !ok || tag != "goat_horn" {
		t.Errorf("GetCooldownTag() = (%q, %v), want (goat_horn, true)", tag, ok)
	}
}

func TestGoatHornStateIdChangesWithType(t *testing.T) {
	g := NewGoatHorn(NewItemIdentifier(GOAT_HORN), "Goat Horn")
	ponder := g.GetStateId()
	g.SetHornType(GoatHornTypeDream)
	if g.GetStateId() == ponder {
		t.Error("expected GetStateId() to change when the horn type changes")
	}
}

func TestSuspiciousStewFoodValues(t *testing.T) {
	s := NewSuspiciousStew(NewItemIdentifier(SUSPICIOUS_STEW), "Suspicious Stew")
	if s.GetFoodRestore() != 6 {
		t.Errorf("GetFoodRestore() = %d, want 6", s.GetFoodRestore())
	}
	if s.GetSaturationRestore() != 7.2 {
		t.Errorf("GetSaturationRestore() = %v, want 7.2", s.GetSaturationRestore())
	}
	if s.RequiresHunger() {
		t.Error("expected SuspiciousStew.RequiresHunger() to be false")
	}
	if s.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", s.GetMaxStackSize())
	}
	if s.GetType() != SuspiciousStewTypePoppy {
		t.Errorf("GetType() = %v, want Poppy", s.GetType())
	}
}

func TestMedicineTypeGetDisplayName(t *testing.T) {
	if got := MedicineTypeEyeDrops.GetDisplayName(); got != "Eye Drops" {
		t.Errorf("GetDisplayName() = %q, want %q", got, "Eye Drops")
	}
}

func TestMedicineDefaultsAndState(t *testing.T) {
	m := NewMedicine(NewItemIdentifier(MEDICINE), "Medicine")
	if m.GetType() != MedicineTypeEyeDrops {
		t.Errorf("GetType() = %v, want EyeDrops", m.GetType())
	}
	if m.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", m.GetMaxStackSize())
	}

	m.SetType(MedicineTypeAntidote)
	if m.GetType() != MedicineTypeAntidote {
		t.Errorf("GetType() = %v, want Antidote", m.GetType())
	}
}
