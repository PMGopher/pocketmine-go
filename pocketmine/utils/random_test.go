package utils

import "testing"

func TestRandomIsDeterministicForSameSeed(t *testing.T) {
	a := NewRandom(12345)
	b := NewRandom(12345)
	for i := 0; i < 100; i++ {
		av, bv := a.NextInt(), b.NextInt()
		if av != bv {
			t.Fatalf("iteration %d: got %d and %d for the same seed", i, av, bv)
		}
		if av < 0 || av > 0x7fffffff {
			t.Fatalf("NextInt() = %d out of 31-bit range", av)
		}
	}
}

func TestRandomNextBoundedInt(t *testing.T) {
	r := NewRandom(1)
	for i := 0; i < 1000; i++ {
		v := r.NextBoundedInt(10)
		if v < 0 || v >= 10 {
			t.Fatalf("NextBoundedInt(10) = %d out of range", v)
		}
	}
}
