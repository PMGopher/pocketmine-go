package noise

import (
	"testing"

	"pocketmine-go/pocketmine/utils"
)

func TestNoise2DIsDeterministicForAFixedSeed(t *testing.T) {
	a := NewSimplex(utils.NewRandom(42), 2, 1.0/16, 1.0/512)
	b := NewSimplex(utils.NewRandom(42), 2, 1.0/16, 1.0/512)

	for _, pos := range [][2]float64{{0, 0}, {100.5, -30.25}, {-999, 12345}} {
		got, want := a.Noise2D(pos[0], pos[1], true), b.Noise2D(pos[0], pos[1], true)
		if got != want {
			t.Errorf("Noise2D(%v) = %v, want %v (same seed must give identical output)", pos, got, want)
		}
	}
}

func TestNoise2DNormalizedStaysWithinUnitRange(t *testing.T) {
	s := NewSimplex(utils.NewRandom(1), 4, 0.5, 1.0/32)
	for x := -50.0; x <= 50; x += 3.7 {
		for z := -50.0; z <= 50; z += 4.3 {
			v := s.Noise2D(x, z, true)
			if v < -1.0001 || v > 1.0001 {
				t.Fatalf("Noise2D(%v,%v, normalized) = %v, want within [-1,1]", x, z, v)
			}
		}
	}
}

func TestNoise3DIsDeterministicForAFixedSeed(t *testing.T) {
	a := NewSimplex(utils.NewRandom(7), 4, 0.25, 1.0/32)
	b := NewSimplex(utils.NewRandom(7), 4, 0.25, 1.0/32)

	got := a.Noise3D(10, 20, 30, false)
	want := b.Noise3D(10, 20, 30, false)
	if got != want {
		t.Errorf("Noise3D = %v, want %v", got, want)
	}
}

func TestGetFastNoise3DMatchesDirectSamplingAtGridPoints(t *testing.T) {
	s := NewSimplex(utils.NewRandom(5), 4, 0.25, 1.0/32)

	xSize, ySize, zSize := 16, 32, 16
	xRate, yRate, zRate := 4, 8, 4
	baseX, baseY, baseZ := 5, -10, 20

	got := s.GetFastNoise3D(xSize, ySize, zSize, xRate, yRate, zRate, baseX, baseY, baseZ)

	for xx := 0; xx <= xSize; xx += xRate {
		for zz := 0; zz <= zSize; zz += zRate {
			for yy := 0; yy <= ySize; yy += yRate {
				want := s.Noise3D(float64(baseX+xx), float64(baseY+yy), float64(baseZ+zz), true)
				if got[xx][zz][yy] != want {
					t.Errorf("GetFastNoise3D[%d][%d][%d] = %v, want %v (exact grid point should match direct sampling)", xx, zz, yy, got[xx][zz][yy], want)
				}
			}
		}
	}
}
