package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

func TestComputeSunAnglePercentageAtNoonAndMidnight(t *testing.T) {
	if got := computeSunAnglePercentage(TimeNoon); got != 0 {
		t.Errorf("computeSunAnglePercentage(TimeNoon) = %v, want 0", got)
	}
	if got := computeSunAnglePercentage(TimeMidnight); got != 0.5 {
		t.Errorf("computeSunAnglePercentage(TimeMidnight) = %v, want 0.5", got)
	}
}

func TestComputeSkyLightReductionAtNoonAndMidnight(t *testing.T) {
	if got := computeSkyLightReduction(computeSunAnglePercentage(TimeNoon)); got != 0 {
		t.Errorf("computeSkyLightReduction at noon = %d, want 0 (full daylight)", got)
	}
	if got := computeSkyLightReduction(computeSunAnglePercentage(TimeMidnight)); got != 11 {
		t.Errorf("computeSkyLightReduction at midnight = %d, want 11 (max reduction)", got)
	}
}

func TestDoTickAdvancesTimeAndRecomputesSunAndSkyLight(t *testing.T) {
	w := newTestWorld()
	w.SetTime(TimeMidnight - 1)

	w.DoTick(1)

	if got := w.GetTime(); got != TimeMidnight {
		t.Errorf("GetTime() after one tick = %d, want %d", got, int64(TimeMidnight))
	}
	if got := w.GetSunAnglePercentage(); got != 0.5 {
		t.Errorf("GetSunAnglePercentage() at midnight = %v, want 0.5", got)
	}
	if got := w.GetSkyLightReduction(); got != 11 {
		t.Errorf("GetSkyLightReduction() at midnight = %d, want 11", got)
	}
}

func TestDoTickDoesNotAdvanceTimeWhenStopped(t *testing.T) {
	w := newTestWorld()
	w.SetTime(100)
	w.StopTime()

	w.DoTick(1)

	if got := w.GetTime(); got != 100 {
		t.Errorf("GetTime() after a tick while stopped = %d, want unchanged 100", got)
	}
	if !w.IsTimeStopped() {
		t.Error("IsTimeStopped() = false after StopTime()")
	}

	w.StartTime()
	w.DoTick(2)
	if got := w.GetTime(); got != 101 {
		t.Errorf("GetTime() after StartTime()+tick = %d, want 101", got)
	}
}

func TestScheduleDelayedBlockUpdateDedupSkipsEqualOrShorterExistingDelay(t *testing.T) {
	w := newTestWorld()
	pos := math.NewVector3(5, 80, 5)

	w.ScheduleDelayedBlockUpdate(pos, 10)
	if len(w.scheduledUpdates) != 1 {
		t.Fatalf("after first schedule: len(scheduledUpdates) = %d, want 1", len(w.scheduledUpdates))
	}

	// A new request with a longer delay than the existing one must be dropped.
	w.ScheduleDelayedBlockUpdate(pos, 20)
	if len(w.scheduledUpdates) != 1 {
		t.Errorf("scheduling a longer delay over an existing shorter one added an entry: len = %d, want 1", len(w.scheduledUpdates))
	}

	// A new request with a shorter delay must replace it (PHP allows a strictly shorter delay
	// through - only delay <= existing skips).
	w.ScheduleDelayedBlockUpdate(pos, 3)
	if len(w.scheduledUpdates) != 2 {
		t.Fatalf("scheduling a shorter delay: len(scheduledUpdates) = %d, want 2 (old entry not removed, just superseded in the index)", len(w.scheduledUpdates))
	}
	if w.scheduledUpdateDelay[[3]int{5, 80, 5}] != 3 {
		t.Errorf("scheduledUpdateDelay after shorter reschedule = %d, want 3", w.scheduledUpdateDelay[[3]int{5, 80, 5}])
	}
}

func TestScheduleDelayedBlockUpdateIgnoresOutOfWorldPositions(t *testing.T) {
	w := newTestWorld()
	w.ScheduleDelayedBlockUpdate(math.NewVector3(0, YMax+1000, 0), 5)
	if len(w.scheduledUpdates) != 0 {
		t.Errorf("scheduling an out-of-world position added an entry: len = %d, want 0", len(w.scheduledUpdates))
	}
}

func TestScheduleDelayedBlockUpdateFiresOnScheduledUpdateWhenDue(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	// Coral.OnNearbyBlockChange breaks itself immediately (via UseBreakOn) if it isn't resting on
	// a supporting block below - see Coral.canBeSupportedAt. SetBlock always notifies neighbours
	// (including itself), so a floating coral would be destroyed the instant it's placed; a solid
	// block underneath keeps this test's coral alive to actually reach OnScheduledUpdate.
	if err := w.SetBlock(block.NewPosition(5, 89, 5, w), block.VanillaStone()); err != nil {
		t.Fatalf("SetBlock (support): %v", err)
	}

	id, err := block.NewBlockIdentifier(9001, nil)
	if err != nil {
		t.Fatal(err)
	}
	coral := block.NewCoral(id, "Test Coral", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))

	pos := block.NewPosition(5, 90, 5, w)
	if err := w.SetBlock(pos, coral); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}

	if dead := w.GetBlockAt(5, 90, 5).(block.CoralMaterial).IsDead(); dead {
		t.Fatal("precondition failed: coral already dead before scheduling")
	}

	w.ScheduleDelayedBlockUpdate(pos.AsVector3(), 5)

	// Not due yet.
	w.DoTick(4)
	if dead := w.GetBlockAt(5, 90, 5).(block.CoralMaterial).IsDead(); dead {
		t.Fatal("coral died before its scheduled update was due")
	}

	// Due now (dueTick = currentTick-at-schedule-time (0) + delay (5) = 5).
	w.DoTick(5)
	if dead := w.GetBlockAt(5, 90, 5).(block.CoralMaterial).IsDead(); !dead {
		t.Error("coral did not die after its scheduled update fired")
	}
}

func TestSetBlockQueuesNeighbourUpdatesForSelfAndSixSides(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(5, 90, 5, w)

	if err := w.SetBlock(pos, block.VanillaStone()); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}

	if got := len(w.neighbourUpdateQueue); got != 7 {
		t.Errorf("len(neighbourUpdateQueue) after SetBlock = %d, want 7 (self + 6 sides)", got)
	}

	w.updateNeighbourBlockUpdates()

	if got := len(w.neighbourUpdateQueue); got != 0 {
		t.Errorf("len(neighbourUpdateQueue) after draining = %d, want 0", got)
	}
	if got := len(w.neighbourUpdateQueued); got != 0 {
		t.Errorf("len(neighbourUpdateQueued) after draining = %d, want 0", got)
	}
}

func TestNotifyNeighbourBlockUpdateDedupsAlreadyQueuedPosition(t *testing.T) {
	w := newTestWorld()
	// Each call queues the position itself plus its 6 neighbours (7 total, see
	// internalNotifyNeighbourBlockUpdate) - calling it twice for the same position must not queue
	// any of those 7 a second time.
	w.NotifyNeighbourBlockUpdate(math.NewVector3(5, 90, 5))
	w.NotifyNeighbourBlockUpdate(math.NewVector3(5, 90, 5))

	if got := len(w.neighbourUpdateQueue); got != 7 {
		t.Errorf("len(neighbourUpdateQueue) after notifying the same position twice = %d, want 7 (no growth on the second call)", got)
	}
}

func TestIsChunkTickableRequiresAllNineChunksGeneratedPopulatedAndLightPopulated(t *testing.T) {
	w := newTestWorld()

	// generateChunkOnly-only via ensurePopulated's neighbour generation: after loading just
	// (0,0), its neighbours exist but were only generated, never populated.
	w.GetOrLoadChunk(0, 0)
	if w.isChunkTickable(0, 0) {
		t.Fatal("isChunkTickable(0,0) = true with unpopulated neighbours, want false")
	}

	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			w.GetOrLoadChunk(dx, dz)
		}
	}
	if !w.isChunkTickable(0, 0) {
		t.Error("isChunkTickable(0,0) = false after every neighbour was fully loaded, want true")
	}
}

func TestTickChunksEventuallyRandomTicksGrassIntoDirtWhenStarvedOfLight(t *testing.T) {
	w := newTestWorld()

	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			w.GetOrLoadChunk(dx, dz)
		}
	}

	// Flat's grass layer sits at a known height (see VanillaFlatLayers) - find it by scanning
	// down from a safely-above-terrain Y for the first Grass block.
	var grassY int = -1
	for y := 0; y < 128; y++ {
		if w.GetBlockAt(5, y, 5).GetTypeId() == block.GRASS {
			grassY = y
			break
		}
	}
	if grassY == -1 {
		t.Fatal("no grass block found in generated Flat terrain - test assumption is wrong")
	}

	// Seal the grass in on all sides except below with stone, so GetFullLightAt above it is
	// starved (< 4) and its light filter is high (>= 2) - Grass.OnRandomTick's exact "grass dies"
	// condition.
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			for dy := 1; dy <= 3; dy++ {
				if err := w.SetBlock(block.NewPosition(float64(5+dx), float64(grassY+dy), float64(5+dz), w), block.VanillaStone()); err != nil {
					t.Fatalf("SetBlock: %v", err)
				}
			}
		}
	}

	// SetBlock only queues light recalculation (see its own doc comment) - it isn't executed until
	// DoTick's end-of-tick Execute() calls, which this test bypasses by calling tickChunks()
	// directly. Without this, GetFullLightAt below would still see the stale, pre-roof light
	// values and never actually starve the grass.
	w.blockLightUpdate.Execute()
	w.skyLightUpdate.Execute()

	w.RegisterTickingChunk("test-loader", 0, 0)

	// tickChunk only samples 3 random positions per non-empty subchunk per call - loop enough
	// times that the probability of never sampling this one position is astronomically small
	// (P(miss) per call <= 1 - 3/4096, so ~5000 calls drives P(always miss) well under 1e-6).
	for i := 0; i < 5000; i++ {
		if w.GetBlockAt(5, grassY, 5).GetTypeId() != block.GRASS {
			break
		}
		w.tickChunks()
	}

	if got := w.GetBlockAt(5, grassY, 5).GetTypeId(); got != block.DIRT {
		t.Errorf("grass starved of light GetTypeId() after many ticks = %d, want DIRT (%d)", got, block.DIRT)
	}
}

func TestRegisterAndUnregisterChunkLoaderQueuesForUnloadOnlyOnceLoaderFree(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(2, 2)

	if w.IsChunkInUse(2, 2) {
		t.Fatal("IsChunkInUse(2,2) = true before any loader registered")
	}

	w.RegisterChunkLoader("loader-a", 2, 2)
	w.RegisterChunkLoader("loader-b", 2, 2)
	if !w.IsChunkInUse(2, 2) {
		t.Fatal("IsChunkInUse(2,2) = false with two loaders registered")
	}

	w.UnregisterChunkLoader("loader-a", 2, 2)
	if !w.IsChunkInUse(2, 2) {
		t.Error("IsChunkInUse(2,2) = false after removing only one of two loaders")
	}
	if _, queued := w.unloadQueue[[2]int{2, 2}]; queued {
		t.Error("chunk was queued for unload while a loader remains")
	}

	w.UnregisterChunkLoader("loader-b", 2, 2)
	if w.IsChunkInUse(2, 2) {
		t.Error("IsChunkInUse(2,2) = true after removing every loader")
	}
	if _, queued := w.unloadQueue[[2]int{2, 2}]; !queued {
		t.Error("chunk was not queued for unload after its last loader was removed")
	}
}

func TestUnloadChunksRespectsGraceWindowAndSkipsInUseChunks(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(3, 3)
	w.GetOrLoadChunk(4, 4)

	w.currentTick = 0
	w.RegisterChunkLoader("only-loader", 3, 3)
	w.UnregisterChunkLoader("only-loader", 3, 3) // queued for unload at tick 0
	w.UnregisterChunkLoader("never-registered", 4, 4)
	w.unloadQueue[[2]int{4, 4}] = 0

	// Re-register (3,3) as in-use again before the grace window elapses - it must survive even
	// once the grace window passes, matching real PHP's own "still in use" re-check.
	w.RegisterChunkLoader("new-loader", 3, 3)

	w.currentTick = unloadGraceTicks
	w.unloadChunks()

	if _, ok := w.GetChunk(3, 3); !ok {
		t.Error("chunk (3,3) was unloaded despite having an active loader")
	}
	if _, ok := w.GetChunk(4, 4); ok {
		t.Error("chunk (4,4) was not unloaded after its grace window elapsed with no loaders")
	}
}

func TestUnloadChunksDoesNothingBeforeGraceWindowElapses(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(6, 6)
	w.currentTick = 0
	w.unloadQueue[[2]int{6, 6}] = 0

	w.currentTick = unloadGraceTicks - 1
	w.unloadChunks()

	if _, ok := w.GetChunk(6, 6); !ok {
		t.Error("chunk (6,6) was unloaded before its grace window elapsed")
	}
}
