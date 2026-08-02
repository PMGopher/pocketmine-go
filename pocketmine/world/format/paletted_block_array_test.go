package format

import "testing"

func TestNewPalettedBlockArrayIsUniform(t *testing.T) {
	p := NewPalettedBlockArray(7)
	if p.GetBitsPerBlock() != 0 {
		t.Errorf("GetBitsPerBlock() = %d, want 0", p.GetBitsPerBlock())
	}
	if got := p.Get(0, 0, 0); got != 7 {
		t.Errorf("Get(0,0,0) = %d, want 7", got)
	}
	if got := p.Get(15, 15, 15); got != 7 {
		t.Errorf("Get(15,15,15) = %d, want 7", got)
	}
	if len(p.GetWordArray()) != 0 {
		t.Errorf("expected an empty word array for a uniform layer, got %d bytes", len(p.GetWordArray()))
	}
}

func TestSetAndGetRoundTripsEveryPosition(t *testing.T) {
	p := NewPalettedBlockArray(0)
	// Fill with a value derived from position so every (x,y,z) is independently checkable.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				p.Set(x, y, z, int32(x*256+y*16+z))
			}
		}
	}
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				want := int32(x*256 + y*16 + z)
				if got := p.Get(x, y, z); got != want {
					t.Fatalf("Get(%d,%d,%d) = %d, want %d", x, y, z, got, want)
				}
			}
		}
	}
}

func TestSetGrowsBitsPerBlockAsPaletteGrows(t *testing.T) {
	p := NewPalettedBlockArray(0)
	if p.GetBitsPerBlock() != 0 {
		t.Fatalf("initial bitsPerBlock = %d, want 0", p.GetBitsPerBlock())
	}

	p.Set(0, 0, 0, 1) // 2 distinct values (0, 1) -> needs 1 bit
	if p.GetBitsPerBlock() != 1 {
		t.Errorf("bitsPerBlock after 2nd distinct value = %d, want 1", p.GetBitsPerBlock())
	}

	p.Set(0, 0, 1, 2) // 3 distinct values -> needs 2 bits
	if p.GetBitsPerBlock() != 2 {
		t.Errorf("bitsPerBlock after 3rd distinct value = %d, want 2", p.GetBitsPerBlock())
	}

	p.Set(0, 0, 2, 3)
	p.Set(0, 0, 3, 4) // 5 distinct values -> needs 3 bits
	if p.GetBitsPerBlock() != 3 {
		t.Errorf("bitsPerBlock after 5th distinct value = %d, want 3", p.GetBitsPerBlock())
	}

	// Values set before each grow must still read back correctly after repacking.
	cases := []struct {
		x, y, z int
		want    int32
	}{
		{0, 0, 0, 1},
		{0, 0, 1, 2},
		{0, 0, 2, 3},
		{0, 0, 3, 4},
		{1, 1, 1, 0}, // never explicitly set - should still read the original fill value
	}
	for _, c := range cases {
		if got := p.Get(c.x, c.y, c.z); got != c.want {
			t.Errorf("Get(%d,%d,%d) = %d, want %d", c.x, c.y, c.z, got, c.want)
		}
	}
}

func TestGetPaletteReturnsACopy(t *testing.T) {
	p := NewPalettedBlockArray(5)
	palette := p.GetPalette()
	palette[0] = 999
	if got := p.Get(0, 0, 0); got != 5 {
		t.Errorf("expected mutating the returned palette slice not to affect the array, got %d", got)
	}
}

func TestCollectGarbageShrinksUnusedPaletteEntries(t *testing.T) {
	p := NewPalettedBlockArray(0)
	p.Set(0, 0, 0, 1)
	p.Set(0, 0, 1, 2)
	// Overwrite every position holding 1 or 2 back to 0, so those palette entries become unused.
	p.Set(0, 0, 0, 0)
	p.Set(0, 0, 1, 0)

	p.CollectGarbage()

	if p.GetBitsPerBlock() != 0 {
		t.Errorf("expected collapse back to uniform (bitsPerBlock 0), got %d", p.GetBitsPerBlock())
	}
	if got := p.Get(5, 5, 5); got != 0 {
		t.Errorf("Get(5,5,5) = %d, want 0", got)
	}
}

func TestCollectGarbageKeepsUsedValuesReadable(t *testing.T) {
	p := NewPalettedBlockArray(0)
	p.Set(0, 0, 0, 10)
	p.Set(1, 0, 0, 20)
	p.Set(2, 0, 0, 30)
	p.Set(0, 0, 0, 0) // 10 is now unused

	p.CollectGarbage()

	if got := p.Get(1, 0, 0); got != 20 {
		t.Errorf("Get(1,0,0) = %d, want 20", got)
	}
	if got := p.Get(2, 0, 0); got != 30 {
		t.Errorf("Get(2,0,0) = %d, want 30", got)
	}
	if got := p.Get(0, 0, 0); got != 0 {
		t.Errorf("Get(0,0,0) = %d, want 0", got)
	}
	if len(p.GetPalette()) != 3 { // 0, 20, 30
		t.Errorf("len(GetPalette()) = %d, want 3", len(p.GetPalette()))
	}
}

func TestBitsPerBlockForOnlyReturnsValidSteps(t *testing.T) {
	cases := []struct {
		paletteSize int
		want        int
	}{
		{1, 0}, {2, 1}, {3, 2}, {4, 2}, {5, 3}, {8, 3}, {9, 4}, {16, 4}, {17, 5}, {64, 6}, {65, 8}, {256, 8}, {257, 16},
	}
	for _, c := range cases {
		if got := bitsPerBlockFor(c.paletteSize); got != c.want {
			t.Errorf("bitsPerBlockFor(%d) = %d, want %d", c.paletteSize, got, c.want)
		}
	}
}

func TestGetWordArrayLengthMatchesBitsPerBlock(t *testing.T) {
	p := NewPalettedBlockArray(0)
	p.Set(0, 0, 0, 1) // bitsPerBlock 1: 32 blocks/word, 4096/32 = 128 words = 512 bytes
	if got := len(p.GetWordArray()); got != 512 {
		t.Errorf("word array length at bitsPerBlock=1 = %d, want 512", got)
	}
}
