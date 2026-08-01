package item

import "testing"

var (
	_ Item = (*Apple)(nil)
	_ Item = (*Bread)(nil)
	_ Item = (*Carrot)(nil)
)

func TestAppleFoodValues(t *testing.T) {
	a := NewApple(NewItemIdentifier(APPLE), "Apple")
	if a.GetFoodRestore() != 4 {
		t.Errorf("GetFoodRestore() = %d, want 4", a.GetFoodRestore())
	}
	if a.GetSaturationRestore() != 2.4 {
		t.Errorf("GetSaturationRestore() = %v, want 2.4", a.GetSaturationRestore())
	}
	if !a.RequiresHunger() {
		t.Error("expected Apple to require hunger")
	}
}

func TestBreadFoodValues(t *testing.T) {
	b := NewBread(NewItemIdentifier(BREAD), "Bread")
	if b.GetFoodRestore() != 5 {
		t.Errorf("GetFoodRestore() = %d, want 5", b.GetFoodRestore())
	}
	if b.GetSaturationRestore() != 6 {
		t.Errorf("GetSaturationRestore() = %v, want 6", b.GetSaturationRestore())
	}
}

func TestCarrotFoodValues(t *testing.T) {
	c := NewCarrot(NewItemIdentifier(CARROT), "Carrot")
	if c.GetFoodRestore() != 3 {
		t.Errorf("GetFoodRestore() = %d, want 3", c.GetFoodRestore())
	}
	if c.GetSaturationRestore() != 4.8 {
		t.Errorf("GetSaturationRestore() = %v, want 4.8", c.GetSaturationRestore())
	}
}

func TestFoodCloneIsIndependent(t *testing.T) {
	a := NewApple(NewItemIdentifier(APPLE), "Apple")
	a.SetCount(3)

	clone := a.Clone().(*Apple)
	clone.SetCount(1)

	if a.GetCount() != 3 {
		t.Errorf("original GetCount() = %d, want 3 (clone mutation leaked)", a.GetCount())
	}
}
