package item

import "testing"

var (
	_ Item = (*Clownfish)(nil)
	_ Item = (*Pufferfish)(nil)
)

func TestClownfishFoodValues(t *testing.T) {
	c := NewClownfish(NewItemIdentifier(CLOWNFISH), "Clownfish")
	if c.GetFoodRestore() != 1 {
		t.Errorf("GetFoodRestore() = %d, want 1", c.GetFoodRestore())
	}
	if c.GetSaturationRestore() != 0.2 {
		t.Errorf("GetSaturationRestore() = %v, want 0.2", c.GetSaturationRestore())
	}
}

func TestPufferfishFoodValues(t *testing.T) {
	p := NewPufferfish(NewItemIdentifier(PUFFERFISH), "Pufferfish")
	if p.GetFoodRestore() != 1 {
		t.Errorf("GetFoodRestore() = %d, want 1", p.GetFoodRestore())
	}
	if p.GetSaturationRestore() != 0.2 {
		t.Errorf("GetSaturationRestore() = %v, want 0.2", p.GetSaturationRestore())
	}
}
