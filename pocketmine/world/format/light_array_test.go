package format

import "testing"

func TestLightArrayFillAndGetSet(t *testing.T) {
	l := NewLightArrayFilled(15)
	if got := l.Get(0, 0, 0); got != 15 {
		t.Errorf("Get(0,0,0) = %d, want 15", got)
	}
	if got := l.Get(15, 15, 15); got != 15 {
		t.Errorf("Get(15,15,15) = %d, want 15", got)
	}

	l.Set(3, 4, 5, 7)
	if got := l.Get(3, 4, 5); got != 7 {
		t.Errorf("Get(3,4,5) after Set = %d, want 7", got)
	}
	// Neighbouring nibble (packed into the same byte) must be untouched.
	if got := l.Get(3, 4, 4); got != 15 {
		t.Errorf("Get(3,4,4) (neighbouring nibble) = %d, want 15 (untouched)", got)
	}
}

func TestLightArraySetEveryPositionRoundTrips(t *testing.T) {
	l := NewLightArrayFilled(0)
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				l.Set(x, y, z, (x+y+z)%16)
			}
		}
	}
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				want := (x + y + z) % 16
				if got := l.Get(x, y, z); got != want {
					t.Fatalf("Get(%d,%d,%d) = %d, want %d", x, y, z, got, want)
				}
			}
		}
	}
}

func TestLightArrayIsUniform(t *testing.T) {
	l := NewLightArrayFilled(5)
	if !l.IsUniform(5) {
		t.Error("expected a freshly filled array to be uniform")
	}
	if l.IsUniform(6) {
		t.Error("expected IsUniform(6) to be false for a value-5 array")
	}
	l.Set(1, 1, 1, 6)
	if l.IsUniform(5) {
		t.Error("expected IsUniform(5) to be false after changing one position")
	}
}

func TestLightArrayCloneIsIndependent(t *testing.T) {
	l := NewLightArrayFilled(3)
	c := l.Clone()
	c.Set(0, 0, 0, 12)
	if got := l.Get(0, 0, 0); got == 12 {
		t.Error("expected Clone to be independent of the original")
	}
}
