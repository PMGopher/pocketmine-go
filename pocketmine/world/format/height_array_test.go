package format

import "testing"

func TestHeightArrayFillAndGetSet(t *testing.T) {
	h := NewHeightArrayFilled(64)
	if got := h.Get(5, 7); got != 64 {
		t.Errorf("Get(5,7) = %d, want 64", got)
	}
	h.Set(5, 7, 100)
	if got := h.Get(5, 7); got != 100 {
		t.Errorf("Get(5,7) after Set = %d, want 100", got)
	}
	if got := h.Get(0, 0); got != 64 {
		t.Errorf("Get(0,0) (untouched) = %d, want 64", got)
	}
}

func TestHeightArrayCloneIsIndependent(t *testing.T) {
	h := NewHeightArrayFilled(10)
	c := h.Clone()
	c.Set(0, 0, 999)
	if got := h.Get(0, 0); got == 999 {
		t.Error("expected Clone to be independent of the original")
	}
}

func TestHeightArraySetValuesRoundTrips(t *testing.T) {
	h := NewHeightArrayFilled(0)
	var values [256]int
	for i := range values {
		values[i] = i
	}
	h.SetValues(values)
	if got := h.GetValues(); got != values {
		t.Error("GetValues after SetValues did not round-trip")
	}
}
