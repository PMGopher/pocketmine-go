package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/biome"
)

func TestGetSetBiomeIDRoundTripsAndGetBiomeResolvesIt(t *testing.T) {
	w := newTestWorld()

	w.SetBiomeID(5, 90, 5, int32(biome.IDHell))
	if got := w.GetBiomeID(5, 90, 5); got != int32(biome.IDHell) {
		t.Errorf("GetBiomeID(5,90,5) = %d, want IDHell (%d)", got, biome.IDHell)
	}
	if got := w.GetBiome(5, 90, 5); got.ID() != biome.IDHell {
		t.Errorf("GetBiome(5,90,5).ID() = %d, want IDHell (%d)", got.ID(), biome.IDHell)
	}
}

func TestSpawnLocationDefaultsToZeroAndRoundTripsThroughSetSpawnLocation(t *testing.T) {
	w := newTestWorld()
	if got := w.GetSpawnLocation(); got != (math.Vector3{}) {
		t.Errorf("GetSpawnLocation() on a fresh World = %v, want the zero vector", got)
	}

	pos := math.NewVector3(16, 70, 16)
	w.SetSpawnLocation(pos)
	if got := w.GetSpawnLocation(); got != pos {
		t.Errorf("GetSpawnLocation() after SetSpawnLocation = %v, want %v", got, pos)
	}
}

func TestIsSpawnChunkChecksThe3x3AreaAroundTheSpawnChunk(t *testing.T) {
	w := newTestWorld()
	w.SetSpawnLocation(math.NewVector3(20, 70, 20)) // chunk (1,1)

	for _, c := range [][2]int{{0, 0}, {1, 0}, {2, 2}, {0, 2}} {
		if !w.IsSpawnChunk(c[0], c[1]) {
			t.Errorf("IsSpawnChunk%v = false, want true (within 1 chunk of spawn chunk (1,1))", c)
		}
	}
	for _, c := range [][2]int{{3, 1}, {1, 3}, {-1, 1}} {
		if w.IsSpawnChunk(c[0], c[1]) {
			t.Errorf("IsSpawnChunk%v = true, want false (outside the 3x3 spawn area)", c)
		}
	}
}

func TestChunkStateQueriesReflectLoadAndPopulation(t *testing.T) {
	w := newTestWorld()

	if w.IsChunkLoaded(0, 0) || w.IsChunkGenerated(0, 0) || w.IsChunkPopulated(0, 0) {
		t.Fatal("a chunk nothing has touched yet reports as loaded/generated/populated")
	}

	w.GetOrLoadChunk(0, 0)

	if !w.IsChunkLoaded(0, 0) {
		t.Error("IsChunkLoaded(0,0) = false after GetOrLoadChunk")
	}
	if !w.IsChunkGenerated(0, 0) {
		t.Error("IsChunkGenerated(0,0) = false after GetOrLoadChunk")
	}
	if !w.IsChunkPopulated(0, 0) {
		t.Error("IsChunkPopulated(0,0) = false after GetOrLoadChunk")
	}

	loaded := w.GetLoadedChunks()
	if _, ok := loaded[[2]int{0, 0}]; !ok {
		t.Error("GetLoadedChunks() doesn't contain the chunk that was just loaded")
	}
}

func TestGetNearestEntityFindsTheClosestMatchIgnoringFilteredAndDeadOnes(t *testing.T) {
	w := newTestWorld()

	bb, err := math.NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	near := newFakeEntity(1, bb)
	near.pos = math.NewVector3(1, 0, 0)
	far := newFakeEntity(2, bb)
	far.pos = math.NewVector3(50, 0, 0)
	dead := newFakeEntity(3, bb)
	dead.pos = math.NewVector3(0.5, 0, 0)
	dead.alive = false

	w.AddEntity(near)
	w.AddEntity(far)
	w.AddEntity(dead)

	got := w.GetNearestEntity(math.Vector3Zero(), 100, false, nil)
	if got == nil || got.(*fakeEntity).id != near.id {
		t.Fatalf("GetNearestEntity found %v, want the near (non-dead) entity", got)
	}

	// includeDead=true should now let the closer dead entity win.
	got = w.GetNearestEntity(math.Vector3Zero(), 100, true, nil)
	if got == nil || got.(*fakeEntity).id != dead.id {
		t.Fatalf("GetNearestEntity(includeDead=true) found %v, want the dead (closer) entity", got)
	}

	// A filter that rejects everything should find nothing.
	got = w.GetNearestEntity(math.Vector3Zero(), 100, false, func(e block.Entity) bool { return false })
	if got != nil {
		t.Errorf("GetNearestEntity with an always-false filter found %v, want nil", got)
	}

	// Nothing within maxDistance should find nothing either.
	got = w.GetNearestEntity(math.Vector3Zero(), 0.1, false, nil)
	if got != nil {
		t.Errorf("GetNearestEntity with too-small a maxDistance found %v, want nil", got)
	}
}
