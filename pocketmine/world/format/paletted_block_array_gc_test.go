package format

import (
	"math/rand"
	"testing"
)

// CollectGarbage drops the palette entries no block uses and keeps every block's value.
func TestPalettedBlockArrayCollectGarbageKeepsBlocks(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	p := NewPalettedBlockArray(0)
	var want [16][16][16]int32
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				v := int32(r.Intn(40))
				p.Set(x, y, z, v)
				want[x][y][z] = v
			}
		}
	}
	// Overwrite every block with value < 20: those palette entries become unused.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				if want[x][y][z] < 20 {
					p.Set(x, y, z, 100)
					want[x][y][z] = 100
				}
			}
		}
	}
	before := len(p.GetPalette())
	p.CollectGarbage()
	if after := len(p.GetPalette()); after != 21 {
		t.Errorf("palette: %d entries before, %d after, want 21 (values 20-39 and 100)", before, after)
	}
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				if got := p.Get(x, y, z); got != want[x][y][z] {
					t.Fatalf("block %d,%d,%d = %d after CollectGarbage, want %d", x, y, z, got, want[x][y][z])
				}
			}
		}
	}
}

func BenchmarkPalettedBlockArrayCollectGarbage(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	p := NewPalettedBlockArray(0)
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				p.Set(x, y, z, int32(r.Intn(12)))
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.CollectGarbage()
	}
}
