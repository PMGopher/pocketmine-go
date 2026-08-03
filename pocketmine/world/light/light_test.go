package light

import (
	"testing"

	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/utils"
)

const (
	testAirState   int32 = 0
	testStoneState int32 = 1
	testGlowState  int32 = 2 // a fake light-emitting block, level 14
)

// fakeWorld is a minimal utils.ChunkSource backed by a plain map - enough to exercise the light
// engine without needing a real *world.World (which would create an import cycle: world imports
// world/light).
type fakeWorld struct {
	chunks map[[2]int]*format.Chunk
}

func newFakeWorld() *fakeWorld { return &fakeWorld{chunks: map[[2]int]*format.Chunk{}} }

func (w *fakeWorld) GetChunk(chunkX, chunkZ int) (*format.Chunk, bool) {
	c, ok := w.chunks[[2]int{chunkX, chunkZ}]
	return c, ok
}

// newFlatTestChunk builds a chunk with a solid stone floor from y=-64 to y=stoneTop (inclusive),
// air everywhere else - similar in shape to this port's Flat generator, but built directly against
// format.Chunk so this package's tests don't need to import world/generator (which itself imports
// world/format, not a cycle risk, but keeping world/light's tests dependency-free of the rest of
// the world package tree keeps the test's intent focused on light behaviour alone).
func newFlatTestChunk(stoneTop int) *format.Chunk {
	c := format.NewChunk(nil, true, testAirState, 0)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := format.MinSubChunkIndex * format.SubChunkEdgeLength; y <= stoneTop; y++ {
				c.SetBlockStateID(x, y, z, testStoneState)
			}
		}
	}
	return c
}

// testLightFilters mirrors how World builds its real LightFilters map: not a block's raw
// GetLightFilter() (0 for transparent blocks like air, 15 for opaque ones), but
// min(15, GetLightFilter() + BaseLightFilter) - see RuntimeBlockStateRegistry.php's
// `$block->getLightFilter() + LightUpdate::BASE_LIGHT_FILTER`. This is what makes light actually
// attenuate by 1 per block travelled through open air (raw filter 0) while still fully blocking at
// opaque blocks (raw filter 15, already capped).
func testLightFilters() map[int32]int {
	return map[int32]int{
		testAirState:   min(15, 0+BaseLightFilter),
		testStoneState: min(15, 15+BaseLightFilter),
		testGlowState:  min(15, 0+BaseLightFilter),
	}
}

func TestSkyLightRecalculateChunkLightsOpenSkyToFull(t *testing.T) {
	w := newFakeWorld()
	chunk := newFlatTestChunk(0)
	w.chunks[[2]int{0, 0}] = chunk

	explorer := utils.NewSubChunkExplorer(w)
	update := NewSkyLightUpdate(explorer, testLightFilters(), map[int32]bool{testStoneState: true})
	update.RecalculateChunk(0, 0)
	update.Execute()

	if got := chunk.GetBlockSkyLightAt(5, 100, 5); got != 15 {
		t.Errorf("sky light at (5,100,5), well above the floor = %d, want 15", got)
	}
	if got := chunk.GetBlockSkyLightAt(5, 0, 5); got != 0 {
		t.Errorf("sky light inside the stone floor (5,0,5) = %d, want 0", got)
	}
	if got := chunk.GetBlockSkyLightAt(5, 1, 5); got != 15 {
		t.Errorf("sky light directly above the floor (5,1,5) = %d, want 15 (open sky)", got)
	}
}

func TestSkyLightDropsUnderAnOverhang(t *testing.T) {
	w := newFakeWorld()
	chunk := newFlatTestChunk(0)
	// A wide stone roof 5 blocks above the floor - wide enough that its centre isn't reachable by
	// the light engine's deliberate "let light leak in near a cliff edge" approximation (a single
	// isolated 1-wide roof tile right next to open sky legitimately lets some side light under it,
	// matching real Minecraft/PocketMine behaviour - see SkyLightUpdate::recalculateChunk's own
	// maxAdjacentHeight/nodeColumnEnd comment - so a real shadow test needs enough width that the
	// centre is insulated from that edge effect).
	for x := 2; x <= 13; x++ {
		for z := 2; z <= 13; z++ {
			chunk.SetBlockStateID(x, 5, z, testStoneState)
		}
	}
	w.chunks[[2]int{0, 0}] = chunk

	explorer := utils.NewSubChunkExplorer(w)
	update := NewSkyLightUpdate(explorer, testLightFilters(), map[int32]bool{testStoneState: true})
	update.RecalculateChunk(0, 0)
	update.Execute()

	if got := chunk.GetBlockSkyLightAt(7, 4, 7); got >= 15 {
		t.Errorf("sky light directly under the roof's centre (7,4,7) = %d, want < 15 (shadowed)", got)
	}
	// A column with no roof, well outside the roofed area, should still be fully lit.
	if got := chunk.GetBlockSkyLightAt(0, 4, 0); got != 15 {
		t.Errorf("sky light in an unshadowed column (0,4,0) = %d, want 15", got)
	}
}

func TestBlockLightPropagatesFromAnEmitterAndFadesWithDistance(t *testing.T) {
	w := newFakeWorld()
	c := format.NewChunk(nil, true, testAirState, 0)
	c.SetBlockStateID(8, 64, 8, testGlowState)
	w.chunks[[2]int{0, 0}] = c

	filters := testLightFilters()
	emitters := map[int32]int{testGlowState: 14}

	explorer := utils.NewSubChunkExplorer(w)
	update := NewBlockLightUpdate(explorer, filters, emitters)
	update.RecalculateChunk(0, 0)
	update.Execute()

	if got := c.GetBlockLightAt(8, 64, 8); got != 14 {
		t.Errorf("light level at the emitter itself = %d, want 14", got)
	}
	if got := c.GetBlockLightAt(9, 64, 8); got != 13 {
		t.Errorf("light level one block away = %d, want 13 (one less, matches BaseLightFilter=1 through air)", got)
	}
	if got := c.GetBlockLightAt(8, 64, 8+6); got != 8 {
		t.Errorf("light level 6 blocks away = %d, want 8", got)
	}
	if got := c.GetBlockLightAt(0, 64, 0); got != 0 {
		t.Errorf("light level far from the emitter (0,64,0) = %d, want 0", got)
	}
}

func TestSkyLightRemovalDarkensAColumnWhenARoofIsAdded(t *testing.T) {
	w := newFakeWorld()
	chunk := newFlatTestChunk(0)
	w.chunks[[2]int{0, 0}] = chunk

	filters := testLightFilters()
	blockers := map[int32]bool{testStoneState: true}
	explorer := utils.NewSubChunkExplorer(w)
	update := NewSkyLightUpdate(explorer, filters, blockers)
	update.RecalculateChunk(0, 0)
	update.Execute()

	if got := chunk.GetBlockSkyLightAt(7, 4, 7); got != 15 {
		t.Fatalf("precondition failed: expected open sky at (7,4,7) before adding a roof, got %d", got)
	}

	// Now place a wide roof (see TestSkyLightDropsUnderAnOverhang on why it must be wide) and ask
	// the light engine to recalculate each newly placed node - matching how World would react to a
	// series of real block placements (RecalculateNode, not a full RecalculateChunk).
	for x := 2; x <= 13; x++ {
		for z := 2; z <= 13; z++ {
			chunk.SetBlockStateID(x, 5, z, testStoneState)
			update.RecalculateNode(x, 5, z)
		}
	}
	update.Execute()

	if got := chunk.GetBlockSkyLightAt(7, 4, 7); got >= 15 {
		t.Errorf("sky light under the newly placed roof's centre (7,4,7) = %d, want < 15", got)
	}
}
