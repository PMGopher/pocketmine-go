package format

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// fakeTile is a minimal tile.Tile for exercising Chunk's tile storage in isolation - the real
// concrete tile types (pocketmine/block/tile) don't need to exist for Chunk's own bookkeeping to
// be tested.
type fakeTile struct {
	pos    tile.Position
	closed bool
}

func newFakeTile(x, y, z float64) *fakeTile {
	return &fakeTile{pos: tile.NewPosition(math.NewVector3(x, y, z), nil)}
}

func (f *fakeTile) ReadSaveData(n *nbt.CompoundTag) error { return nil }
func (f *fakeTile) WriteSaveData(n *nbt.CompoundTag)      {}
func (f *fakeTile) SaveID() string                        { return "Fake" }
func (f *fakeTile) GetPosition() tile.Position            { return f.pos }
func (f *fakeTile) IsClosed() bool                        { return f.closed }
func (f *fakeTile) Close()                                { f.closed = true }
func (f *fakeTile) OnBlockDestroyed()                     {}
func (f *fakeTile) CopyDataFromItem(item tile.Item)       {}

func TestChunkCloneIsIndependentOfOriginal(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	c.SetBlockStateID(0, 0, 0, 5)

	clone := c.Clone()
	clone.SetBlockStateID(0, 0, 0, 99)

	if got := c.GetBlockStateID(0, 0, 0); got != 5 {
		t.Errorf("expected mutating the clone not to affect the original, got %d", got)
	}
	if got := clone.GetBlockStateID(0, 0, 0); got != 99 {
		t.Errorf("GetBlockStateID(0,0,0) on the clone = %d, want 99", got)
	}
}

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

func TestChunkAddGetRemoveTile(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	tl := newFakeTile(3, 70, 9)

	c.AddTile(tl)
	got, ok := c.GetTile(3, 70, 9)
	if !ok || got != tl {
		t.Fatalf("GetTile after AddTile = (%v, %v), want (tl, true)", got, ok)
	}

	c.RemoveTile(tl)
	if _, ok := c.GetTile(3, 70, 9); ok {
		t.Error("expected GetTile to report not-found after RemoveTile")
	}
}

func TestChunkAddTilePanicsOnClosedTile(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	tl := newFakeTile(0, 0, 0)
	tl.Close()

	defer func() {
		if recover() == nil {
			t.Error("expected AddTile to panic for a closed tile")
		}
	}()
	c.AddTile(tl)
}

func TestChunkOnUnloadClosesEveryTile(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	a, b := newFakeTile(0, 0, 0), newFakeTile(1, 1, 1)
	c.AddTile(a)
	c.AddTile(b)

	c.OnUnload()

	if !a.IsClosed() || !b.IsClosed() {
		t.Error("expected OnUnload to close every tile in the chunk")
	}
}

func TestChunkHeightMapDefaultsToTop(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	want := (MaxSubChunkIndex + 1) * SubChunkEdgeLength
	if got := c.GetHeightMap(5, 5); got != want {
		t.Errorf("default heightmap value = %d, want %d", got, want)
	}
	c.SetHeightMap(5, 5, 64)
	if got := c.GetHeightMap(5, 5); got != 64 {
		t.Errorf("GetHeightMap after Set = %d, want 64", got)
	}
}

func TestChunkTerrainDirtyFlags(t *testing.T) {
	c := NewChunk(nil, false, 0, 1)
	if !c.IsTerrainDirty() {
		t.Error("expected a freshly constructed chunk to start dirty (DirtyFlagsAll)")
	}
	c.ClearTerrainDirtyFlags()
	if c.IsTerrainDirty() {
		t.Error("expected ClearTerrainDirtyFlags to clear dirtiness")
	}
	c.SetBlockStateID(0, 0, 0, 1)
	if !c.GetTerrainDirtyFlag(DirtyFlagBlocks) {
		t.Error("expected SetBlockStateID to set DirtyFlagBlocks")
	}
	if c.GetTerrainDirtyFlag(DirtyFlagBiomes) {
		t.Error("expected DirtyFlagBiomes to still be clear")
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
