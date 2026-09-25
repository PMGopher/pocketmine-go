package world

import (
	"testing"
	"time"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator"
)

func normalWorldKnownBlocks() []block.Behavior {
	return []block.Behavior{block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(), block.VanillaGravel(), block.VanillaCoalOre(), block.VanillaIronOre(), block.VanillaRedstoneOre(), block.VanillaLapisLazuliOre(), block.VanillaGoldOre(), block.VanillaDiamondOre(), block.VanillaEmeraldOre(), block.VanillaWater(), block.VanillaSand(), block.VanillaSandstone(), block.VanillaSnowLayer(), block.VanillaTallGrass(), block.VanillaOakLog(), block.VanillaOakLeaves(), block.VanillaSpruceLog(), block.VanillaSpruceLeaves(), block.VanillaBirchLog(), block.VanillaBirchLeaves()}
}

// newAsyncWorld is a Normal world populating on a real AsyncPool (like the server's).
func newAsyncWorld(t *testing.T) (*World, *scheduler.AsyncPool) {
	t.Helper()
	pool := scheduler.NewAsyncPool(2, log.NewSimpleLogger())
	t.Cleanup(pool.Shutdown)
	w := New(generator.NewNormal(1), convert.NewBlockTranslator(), normalWorldKnownBlocks())
	w.SetGeneratorExecutor(NewAsyncGeneratorExecutor(log.NewSimpleLogger(), pool, GeneratorExecutorSetupParameters{
		WorldMinY:    YMin,
		WorldMaxY:    YMax,
		NewGenerator: func() (generator.Generator, error) { return generator.NewNormal(1), nil },
	}), pool)
	return w, pool
}

// collectUntil collects the pool's results (the server does it every tick) until done is true.
func collectUntil(t *testing.T, pool *scheduler.AsyncPool, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !done() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for async population")
		}
		if _, err := pool.CollectTasks(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestRequestChunkPopulationRunsOnWorkers(t *testing.T) {
	w, pool := newAsyncWorld(t)
	loader := &temporaryChunkLoader{}
	w.RegisterChunkLoader(loader, 3, 3)

	var populated *format.Chunk
	p := w.RequestChunkPopulation(3, 3, loader)
	p.OnCompletion(func(c *format.Chunk) { populated = c }, func() { t.Error("population was rejected") })
	if populated != nil {
		t.Fatal("population of an ungenerated chunk shouldn't complete synchronously")
	}
	if !w.IsChunkLocked(3, 3) || !w.IsChunkLocked(4, 4) {
		t.Error("the chunk and its neighbours should be locked while population runs")
	}

	collectUntil(t, pool, func() bool { return populated != nil })
	if !populated.IsPopulated() {
		t.Error("the resolved chunk isn't populated")
	}
	if c, ok := w.GetChunk(3, 3); !ok || c != populated {
		t.Error("the populated chunk wasn't set in the world")
	}
	if w.IsChunkLocked(3, 3) || len(w.activeChunkPopulationTasks) != 0 {
		t.Error("locks/active tasks weren't released")
	}
	if _, ok := w.GetChunk(2, 2); !ok {
		t.Error("the generated neighbours weren't set in the world")
	}
	if len(w.changedBlocks) != 0 {
		t.Error("population writes were tracked as block changes")
	}

	// Already populated: the promise resolves right away.
	again := false
	w.RequestChunkPopulation(3, 3, loader).OnCompletion(func(*format.Chunk) { again = true }, func() {})
	if !again {
		t.Error("requesting a populated chunk should resolve immediately")
	}
}

func TestModifyingALockedChunkDiscardsThePopulationResult(t *testing.T) {
	w, pool := newAsyncWorld(t)
	loader := &temporaryChunkLoader{}
	w.RegisterChunkLoader(loader, 0, 0)
	w.GetOrLoadChunk(5, 5) // a populated chunk next to one that gets populated asynchronously

	var populated *format.Chunk
	w.RequestChunkPopulation(0, 0, loader).OnCompletion(func(c *format.Chunk) { populated = c }, func() {})
	// Setting a block in an unlocked chunk doesn't affect the population.
	if err := w.SetBlock(block.NewPosition(80, 100, 80, w), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}
	if w.IsChunkLocked(5, 5) {
		t.Fatal("chunk 5,5 isn't part of 0,0's population")
	}
	// A main-thread modification of a locked chunk breaks its lock (setBlockAt does
	// UnlockChunk(nil)): the result is discarded and population is redone.
	w.UnlockChunk(1, 1, nil)

	collectUntil(t, pool, func() bool { return populated != nil })
	if !populated.IsPopulated() {
		t.Error("the retried population didn't populate the chunk")
	}
}

func TestPopulatedChunkGetsLightWhenTicked(t *testing.T) {
	w, pool := newAsyncWorld(t)
	loader := &temporaryChunkLoader{}
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			w.RegisterChunkLoader(loader, x, z)
		}
	}
	done := 0
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			w.RequestChunkPopulation(x, z, loader).OnCompletion(func(*format.Chunk) { done++ }, func() {})
		}
	}
	collectUntil(t, pool, func() bool { return done == 9 })

	if w.isChunkTickable(0, 0) {
		t.Fatal("chunks without light shouldn't be tickable")
	}
	// The server rechecks tickability every tick, ordering light for one more neighbour each time.
	collectUntil(t, pool, func() bool {
		w.isChunkTickable(0, 0)
		for x := -1; x <= 1; x++ {
			for z := -1; z <= 1; z++ {
				c, _ := w.GetChunk(x, z)
				if lit, known := c.IsLightPopulated(); !known || !lit {
					return false
				}
			}
		}
		return true
	})
	if !w.isChunkTickable(0, 0) {
		t.Error("the lit chunk should now be tickable")
	}
	if got := w.GetFullLightAt(0, 200, 0); got != 15 {
		t.Errorf("sky light high above ground = %d, want 15", got)
	}
}

