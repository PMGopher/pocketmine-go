package format

import "testing"

func TestChunkGetBlockStateIDDefaultsToEmpty(t *testing.T) {
	c := NewChunk(nil, false, 42, 1)
	if got := c.GetBlockStateID(0, 0, 0); got != 42 {
		t.Errorf("GetBlockStateID(0,0,0) = %d, want 42 (the empty/default state ID)", got)
	}
	if got := c.GetBlockStateID(5, -60, 5); got != 42 { // near the world floor (MinSubChunkIndex)
		t.Errorf("GetBlockStateID near the floor = %d, want 42", got)
	}
}

func TestChunkSetAndGetBlockStateID(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	c.SetBlockStateID(3, 70, 9, 99)
	if got := c.GetBlockStateID(3, 70, 9); got != 99 {
		t.Errorf("GetBlockStateID(3,70,9) = %d, want 99", got)
	}
	// A neighbouring Y within a different subchunk must be unaffected.
	if got := c.GetBlockStateID(3, 71, 9); got != 0 {
		t.Errorf("GetBlockStateID(3,71,9) = %d, want unchanged 0", got)
	}
}

func TestChunkGetSubChunkPanicsOutOfRange(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	defer func() {
		if recover() == nil {
			t.Error("expected GetSubChunk to panic for an out-of-range Y")
		}
	}()
	c.GetSubChunk(MaxSubChunkIndex + 1)
}

func TestChunkGetHighestBlockAtFindsTopmostNonEmpty(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	c.SetBlockStateID(0, 5, 0, 1)
	c.SetBlockStateID(0, 40, 0, 1)
	c.SetBlockStateID(0, 100, 0, 1)

	height, ok := c.GetHighestBlockAt(0, 0)
	if !ok {
		t.Fatal("expected GetHighestBlockAt to find a block")
	}
	if height != 100 {
		t.Errorf("GetHighestBlockAt(0,0) = %d, want 100", height)
	}
}

func TestChunkGetHighestBlockAtReportsNotFoundWhenEmpty(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	if _, ok := c.GetHighestBlockAt(0, 0); ok {
		t.Error("expected GetHighestBlockAt to report not-found for an all-empty column")
	}
}

func TestChunkSetAndGetBiomeID(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	if got := c.GetBiomeID(0, 0, 0); got != 1 {
		t.Errorf("default biome = %d, want 1", got)
	}
	c.SetBiomeID(0, 0, 0, 4)
	if got := c.GetBiomeID(0, 0, 0); got != 4 {
		t.Errorf("GetBiomeID after Set = %d, want 4", got)
	}
}

func TestChunkPopulatedFlag(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	if c.IsPopulated() {
		t.Error("expected a freshly constructed chunk to be unpopulated")
	}
	c.SetPopulated(true)
	if !c.IsPopulated() {
		t.Error("expected SetPopulated(true) to take effect")
	}
}

func TestChunkUsesProvidedSubChunks(t *testing.T) {
	custom := NewSubChunk(0, nil, NewPalettedBlockArray(1))
	custom.SetBlockStateID(0, 0, 0, 77)
	c := NewChunk(map[int]*SubChunk{3: custom}, false, 0, 1)

	if got := c.GetBlockStateID(0, 3<<SubChunkCoordBitSize, 0); got != 77 {
		t.Errorf("expected the provided subchunk to be used, got %d", got)
	}
}

func TestSubChunkIsEmptyFastAndSetGrowsLayer(t *testing.T) {
	s := NewSubChunk(0, nil, NewPalettedBlockArray(1))
	if !s.IsEmptyFast() {
		t.Error("expected a fresh subchunk with no layers to report empty")
	}
	s.SetBlockStateID(0, 0, 0, 5)
	if s.IsEmptyFast() {
		t.Error("expected SetBlockStateID to create a layer, making the subchunk non-empty")
	}
	if got := s.GetBlockStateID(0, 0, 0); got != 5 {
		t.Errorf("GetBlockStateID(0,0,0) = %d, want 5", got)
	}
}

func TestSubChunkCollectGarbageDropsRedundantEmptyLayer(t *testing.T) {
	s := NewSubChunk(0, nil, NewPalettedBlockArray(1))
	s.SetBlockStateID(0, 0, 0, 5)
	s.SetBlockStateID(0, 0, 0, 0) // back to the empty value everywhere

	s.CollectGarbage()

	if !s.IsEmptyFast() {
		t.Error("expected CollectGarbage to drop a layer that collapsed back to the empty value")
	}
}
