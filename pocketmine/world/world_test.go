package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/generator"
)

// fakeEntity is a minimal registeredEntity for exercising World's entity registry in isolation -
// pocketmine/entity has no concrete spawnable type yet (just the Entity/Living markers), so tests
// here can't use a real one.
type fakeEntity struct {
	id     int
	closed bool
	bb     math.AxisAlignedBB
	pos    math.Vector3
}

func newFakeEntity(id int, bb math.AxisAlignedBB) *fakeEntity { return &fakeEntity{id: id, bb: bb} }

func (f *fakeEntity) ResetFallDistance()                   {}
func (f *fakeEntity) GetPosition() math.Vector3            { return f.pos }
func (f *fakeEntity) SetOnGround(onGround bool)            {}
func (f *fakeEntity) GetFallDistance() float64             { return 0 }
func (f *fakeEntity) SetFallDistance(fallDistance float64) {}
func (f *fakeEntity) GetBoundingBox() math.AxisAlignedBB   { return f.bb }
func (f *fakeEntity) GetMotion() math.Vector3              { return math.Vector3{} }
func (f *fakeEntity) SetOnFire(seconds int)                {}
func (f *fakeEntity) IsOnFire() bool                       { return false }
func (f *fakeEntity) Extinguish()                          {}
func (f *fakeEntity) ExtinguishWithCause(cause int)        {}
func (f *fakeEntity) CanBeMovedByCurrents() bool           { return true }
func (f *fakeEntity) Attack(source entity.DamageSource)    {}
func (f *fakeEntity) GetID() int                           { return f.id }
func (f *fakeEntity) IsClosed() bool                       { return f.closed }

func newTestWorld() *World {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return New(gen, tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
	})
}

func TestGetBlockAtReadsGeneratedTerrain(t *testing.T) {
	w := newTestWorld()

	got := w.GetBlockAt(5, 0, 5)
	if got.GetTypeId() != block.BEDROCK {
		t.Errorf("GetBlockAt(5,0,5).GetTypeId() = %d, want BEDROCK (%d)", got.GetTypeId(), block.BEDROCK)
	}

	got = w.GetBlockAt(5, 100, 5)
	if got.GetTypeId() != block.AIR {
		t.Errorf("GetBlockAt(5,100,5).GetTypeId() = %d, want AIR (%d)", got.GetTypeId(), block.AIR)
	}
}

func TestGetBlockAtReturnsAPositionedClone(t *testing.T) {
	w := newTestWorld()

	a := w.GetBlockAt(5, 0, 5)
	b := w.GetBlockAt(5, 0, 5)
	if a == b {
		t.Error("expected two GetBlockAt calls to return distinct instances")
	}
	if pos := a.GetPosition(); pos.FloorX() != 5 || pos.FloorY() != 0 || pos.FloorZ() != 5 {
		t.Errorf("GetPosition() = (%d,%d,%d), want (5,0,5)", pos.FloorX(), pos.FloorY(), pos.FloorZ())
	}
}

func TestSetBlockThenGetBlockAtRoundTrips(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(3, 50, 3, w)

	if err := w.SetBlock(pos, block.VanillaStone()); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}
	got := w.GetBlockAt(3, 50, 3)
	if got.GetTypeId() != block.STONE {
		t.Errorf("GetBlockAt after SetBlock = %d, want STONE (%d)", got.GetTypeId(), block.STONE)
	}
}

func TestSetBlockRegistersUnknownBlockAsATemplate(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(0, 10, 0, w)

	// Grass wasn't in newTestWorld's knownBlocks... it actually is; use a block never registered
	// up front to prove SetBlock's self-registration, not New's.
	glass := block.NewGlass(mustTestBlockIdentifier(block.GLASS), "Test Glass", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))
	if err := w.SetBlock(pos, glass); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}

	got := w.GetBlockAt(0, 10, 0)
	if got.GetTypeId() != block.GLASS {
		t.Errorf("GetBlockAt after SetBlock(unregistered type) = %d, want GLASS (%d)", got.GetTypeId(), block.GLASS)
	}
}

func mustTestBlockIdentifier(typeID int) *block.BlockIdentifier {
	idInfo, err := block.NewBlockIdentifier(typeID, nil)
	if err != nil {
		panic(err)
	}
	return idInfo
}

func TestGetOrLoadChunkGeneratesOnFirstAccessAndCaches(t *testing.T) {
	w := newTestWorld()

	first := w.GetOrLoadChunk(0, 0)
	second := w.GetOrLoadChunk(0, 0)
	if first != second {
		t.Error("expected repeated GetOrLoadChunk calls for the same coordinates to return the same chunk")
	}
}

func TestIsInWorldRespectsVerticalBounds(t *testing.T) {
	w := newTestWorld()
	if !w.IsInWorld(0, 0, 0) {
		t.Error("expected Y=0 to be in world")
	}
	if w.IsInWorld(0, YMin-1, 0) {
		t.Error("expected below YMin to be out of world")
	}
	if w.IsInWorld(0, YMax, 0) {
		t.Error("expected YMax itself to be out of world (exclusive upper bound)")
	}
}

func TestUseBreakOnReplacesBlockWithAir(t *testing.T) {
	w := newTestWorld()
	if got := w.GetBlockAt(5, 0, 5); got.GetTypeId() != block.BEDROCK {
		t.Fatalf("precondition failed: expected bedrock at (5,0,5), got %d", got.GetTypeId())
	}

	if !w.UseBreakOn(block.NewPosition(5, 0, 5, w).AsVector3()) {
		t.Fatal("expected UseBreakOn to report success")
	}
	if got := w.GetBlockAt(5, 0, 5); got.GetTypeId() != block.AIR {
		t.Errorf("GetBlockAt after UseBreakOn = %d, want AIR (%d)", got.GetTypeId(), block.AIR)
	}
}

// TestPopulatingAChunkDoesNotCascadeToTheWholeWorld guards against a regression where populating
// chunk (0,0) - which can write a handful of blocks across its border into a neighbour - used to
// recursively populate that neighbour too, whose own population could reach its own neighbour, and
// so on: an unbounded chain reaction that eagerly generated and populated the entire world on a
// single GetOrLoadChunk call. See World.ensurePopulated's doc comment for the fix (generating a
// neighbour and populating a neighbour are kept as two distinct steps).
func TestPopulatingAChunkDoesNotCascadeToTheWholeWorld(t *testing.T) {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), generator.VanillaFlatDecorationPopulators())
	w := New(gen, tr, []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
		block.VanillaGravel(), block.VanillaCoalOre(), block.VanillaIronOre(), block.VanillaRedstoneOre(),
		block.VanillaLapisLazuliOre(), block.VanillaGoldOre(), block.VanillaDiamondOre(),
	})

	w.GetOrLoadChunk(0, 0)

	// Only (0,0) and its 8 immediate neighbours may have been generated - nothing further out.
	if got := len(w.chunks); got > 9 {
		t.Fatalf("len(w.chunks) = %d after loading a single chunk, want <= 9 (population cascaded further than the immediate neighbourhood)", got)
	}
}

// TestNormalGeneratorPopulatesRealGroundCoverWithoutCascading exercises the Normal generator
// end-to-end through World - real noise terrain, real biome-driven GroundCover/Ore/TallGrass
// populators - covering both terrain-shape sanity (Normal has its own more detailed tests for
// that) and the same population-cascade guard as
// TestPopulatingAChunkDoesNotCascadeToTheWholeWorld, since Normal's populators can also write
// across chunk borders (Ore's blast radius, same as Flat's).
func TestNormalGeneratorPopulatesRealGroundCoverWithoutCascading(t *testing.T) {
	tr := convert.NewBlockTranslator()
	gen := generator.NewNormal(2024)
	w := New(gen, tr, []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
		block.VanillaGravel(), block.VanillaCoalOre(), block.VanillaIronOre(), block.VanillaRedstoneOre(),
		block.VanillaLapisLazuliOre(), block.VanillaGoldOre(), block.VanillaDiamondOre(), block.VanillaEmeraldOre(),
		block.VanillaWater(), block.VanillaSand(), block.VanillaSandstone(), block.VanillaSnowLayer(), block.VanillaTallGrass(),
	})

	w.GetOrLoadChunk(0, 0)

	if got := len(w.chunks); got > 9 {
		t.Fatalf("len(w.chunks) after loading a single chunk = %d, want <= 9 (population cascaded further than the immediate neighbourhood)", got)
	}

	// GroundCover should have turned at least some exposed stone into a real biome-appropriate
	// surface block (grass, sand, snow, ...) somewhere in the chunk - if every column were still
	// bare stone/water, GroundCover silently did nothing.
	coveredFound := false
	for x := 0; x < 16 && !coveredFound; x++ {
		for z := 0; z < 16 && !coveredFound; z++ {
			for y := 55; y < 90 && !coveredFound; y++ {
				switch w.GetBlockAt(x, y, z).GetTypeId() {
				case block.GRASS, block.SAND, block.SNOW_LAYER, block.GRAVEL, block.DIRT:
					coveredFound = true
				}
			}
		}
	}
	if !coveredFound {
		t.Error("expected GroundCover to have replaced some surface stone with a biome-appropriate block")
	}
}

// TestLevelDBRoundTripPreservesGeneratedTerrain saves a Normal-generated world to a real LevelDB
// database, then opens a *fresh* World (different generator seed, so if loading silently fell
// back to regenerating instead of reading the saved data, the terrain would come out completely
// different) against the same directory and confirms every sampled block matches exactly.
func TestLevelDBRoundTripPreservesGeneratedTerrain(t *testing.T) {
	dir := t.TempDir()

	knownBlocks := []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
		block.VanillaGravel(), block.VanillaCoalOre(), block.VanillaIronOre(), block.VanillaRedstoneOre(),
		block.VanillaLapisLazuliOre(), block.VanillaGoldOre(), block.VanillaDiamondOre(), block.VanillaEmeraldOre(),
		block.VanillaWater(), block.VanillaSand(), block.VanillaSandstone(), block.VanillaSnowLayer(), block.VanillaTallGrass(),
		block.VanillaOakLog(), block.VanillaOakLeaves(), block.VanillaSpruceLog(), block.VanillaSpruceLeaves(),
		block.VanillaBirchLog(), block.VanillaBirchLeaves(),
	}

	tr := convert.NewBlockTranslator()
	w := New(generator.NewNormal(2468), tr, knownBlocks)
	if err := w.OpenProvider(dir); err != nil {
		t.Fatalf("OpenProvider: %v", err)
	}

	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			w.GetOrLoadChunk(x, z)
		}
	}

	type sample struct{ x, y, z, typeID int }
	var before []sample
	for x := -16; x < 16; x += 2 {
		for z := -16; z < 16; z += 2 {
			for y := 0; y < 90; y += 5 {
				before = append(before, sample{x, y, z, w.GetBlockAt(x, y, z).GetTypeId()})
			}
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	tr2 := convert.NewBlockTranslator()
	w2 := New(generator.NewNormal(13579), tr2, knownBlocks) // different seed on purpose
	if err := w2.OpenProvider(dir); err != nil {
		t.Fatalf("OpenProvider (reload): %v", err)
	}
	defer w2.Close()

	for _, s := range before {
		if got := w2.GetBlockAt(s.x, s.y, s.z).GetTypeId(); got != s.typeID {
			t.Errorf("GetBlockAt(%d,%d,%d) after reload = %d, want %d (loaded from disk)", s.x, s.y, s.z, got, s.typeID)
		}
	}
}

// TestLightIsRealAfterGeneration exercises the real light engine now wired into World (see
// ensurePopulated) against a plain Flat world - open sky above the flat layer stack should be
// fully lit, and inside the topmost opaque block itself should be dark.
func TestLightIsRealAfterGeneration(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	if got := w.GetRealBlockSkyLightAt(5, 100, 5); got != 15 {
		t.Errorf("GetRealBlockSkyLightAt(5,100,5) (well above the grass top) = %d, want 15", got)
	}
	if got := w.GetRealBlockSkyLightAt(5, 63, 5); got != 0 {
		t.Errorf("GetRealBlockSkyLightAt(5,63,5) (the grass block itself, opaque) = %d, want 0", got)
	}
	if got := w.GetFullLightAt(5, 100, 5); got != 15 {
		t.Errorf("GetFullLightAt(5,100,5) = %d, want 15", got)
	}
	if got := w.GetHighestAdjacentFullLightAt(5, 63, 5); got != 15 {
		t.Errorf("GetHighestAdjacentFullLightAt(5,63,5) = %d, want 15 (the open-sky block just above is a neighbour)", got)
	}
}

// TestTileAddGetRemoveRoundTrips exercises World's real tile storage (see format.Chunk's own
// AddTile/GetTile/RemoveTile) using a real concrete tile type from pocketmine/block/tile -
// World now satisfies tile.World (GetTileAt/RemoveTile), which NewEnchantTable needs.
func TestTileAddGetRemoveRoundTrips(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	et := tile.NewEnchantTable(w, math.NewVector3(3, 70, 9))
	w.AddTile(et)

	pos := block.NewPosition(3, 70, 9, w)
	got, ok := w.GetTile(pos)
	if !ok || got != tile.Tile(et) {
		t.Fatalf("GetTile after AddTile = (%v, %v), want (et, true)", got, ok)
	}

	w.RemoveTile(et)
	if _, ok := w.GetTile(pos); ok {
		t.Error("expected GetTile to report not-found after RemoveTile")
	}
}

func TestAddGetRemoveEntity(t *testing.T) {
	w := newTestWorld()
	bb, err := math.NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	if err != nil {
		t.Fatalf("NewAxisAlignedBB: %v", err)
	}
	e := newFakeEntity(1, bb)

	w.AddEntity(e)
	got, ok := w.GetEntity(1)
	if !ok || got != e {
		t.Fatalf("GetEntity(1) = (%v, %v), want (e, true)", got, ok)
	}

	w.RemoveEntity(e)
	if _, ok := w.GetEntity(1); ok {
		t.Error("expected GetEntity to report not-found after RemoveEntity")
	}
}

func TestAddEntityPanicsOnClosedEntity(t *testing.T) {
	w := newTestWorld()
	bb, _ := math.NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	e := newFakeEntity(1, bb)
	e.closed = true

	defer func() {
		if recover() == nil {
			t.Error("expected AddEntity to panic for a closed entity")
		}
	}()
	w.AddEntity(e)
}

func TestGetNearbyEntitiesFindsOverlappingBoundingBoxesOnly(t *testing.T) {
	w := newTestWorld()
	nearBB, _ := math.NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	farBB, _ := math.NewAxisAlignedBB(100, 100, 100, 101, 101, 101)
	near := newFakeEntity(1, nearBB)
	far := newFakeEntity(2, farBB)
	w.AddEntity(near)
	w.AddEntity(far)

	queryBB, _ := math.NewAxisAlignedBB(0.5, 0.5, 0.5, 1.5, 1.5, 1.5)
	nearby := w.GetNearbyEntities(queryBB)

	if len(nearby) != 1 || nearby[0] != block.Entity(near) {
		t.Errorf("GetNearbyEntities = %v, want just [near]", nearby)
	}
}

func TestGetOrLoadChunkAtPositionSetsBlockStateID(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(1, 0, 1, w)

	c, ok := w.GetOrLoadChunkAtPosition(pos)
	if !ok {
		t.Fatal("expected GetOrLoadChunkAtPosition to report success")
	}
	c.SetBlockStateID(1, 200, 1, int(block.VanillaStone().GetStateId()))

	if got := w.GetBlockAt(1, 200, 1); got.GetTypeId() != block.STONE {
		t.Errorf("GetTypeId() after chunk adapter Set = %d, want STONE (%d)", got.GetTypeId(), block.STONE)
	}
}

func TestGetSafeSpawnFindsTheFlatWorldsGroundFromHighUpInTheAir(t *testing.T) {
	w := newTestWorld()
	// newTestWorld's Flat layout (see its own doc comment on generator.VanillaFlatLayers): grass
	// tops out at y=63, air above.
	got := w.GetSafeSpawn(math.NewVector3(5, 100, 5))
	want := math.NewVector3(5, 64, 5)
	if got != want {
		t.Errorf("GetSafeSpawn(5,100,5) = %v, want %v (standing on top of the grass at y=63)", got, want)
	}
}

func TestGetSafeSpawnFindsGroundWhenStartingBuriedInsideSolidStone(t *testing.T) {
	w := newTestWorld()
	got := w.GetSafeSpawn(math.NewVector3(5, 30, 5))
	if got.Y <= 30 {
		t.Errorf("GetSafeSpawn(5,30,5) (buried in stone) = %v, want a Y above the stone/dirt/grass column (> 30)", got)
	}
	if got.X != 5 || got.Z != 5 {
		t.Errorf("GetSafeSpawn(5,30,5) changed the x/z column to %v,%v, want it to stay 5,5", got.X, got.Z)
	}
}
