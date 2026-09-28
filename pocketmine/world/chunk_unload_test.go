package world

import (
	"pocketmine-go/pocketmine/block"
	"testing"
	"time"
)

// Chunks a loader used (and their neighbours loaded for population) must be unloaded once no
// loader uses them and the grace period (30 s) has passed, like World::unloadChunks.
func TestUnusedChunksAreUnloaded(t *testing.T) {
	w := newTestWorld()
	loader := &struct{ name string }{"player"}
	var chunks [][2]int
	for x := 40; x < 50; x++ {
		for z := 40; z < 50; z++ {
			w.RegisterChunkLoader(loader, x, z)
			w.RequestChunkPopulation(x, z, loader)
			chunks = append(chunks, [2]int{x, z})
		}
	}
	tick := int64(1)
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		w.DoTick(tick)
		tick++
		done := true
		for _, c := range chunks {
			if _, ok := w.GetLoadedChunks()[c]; !ok {
				done = false
			}
		}
		if done {
			break
		}
		time.Sleep(time.Millisecond)
	}
	loaded := len(w.GetLoadedChunks())
	for _, c := range chunks {
		w.UnregisterChunkLoader(loader, c[0], c[1])
	}
	for i := 0; i < unloadGraceTicks+100; i++ {
		w.DoTick(tick)
		tick++
	}
	t.Logf("loaded while used: %d, after unload: %d, queue: %d", loaded, len(w.GetLoadedChunks()), len(w.unloadQueue))
	if left := len(w.GetLoadedChunks()); left > 0 {
		kept := 0
		for key := range w.GetLoadedChunks() {
			if _, queued := w.unloadQueue[key]; !queued {
				kept++
			}
		}
		t.Errorf("%d chunks still loaded (%d of them not even queued for unloading)", left, kept)
	}
}

// World::getBlockAt returns air outside the world's height instead of reading a chunk (clicking at
// the top or bottom of the world crashed the server through syncBlocksNearby).
func TestGetBlockAtOutsideTheWorldIsAir(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	for _, y := range []int{YMin - 1, YMin - 50, YMax, YMax + 10} {
		if w.GetBlockAt(3, y, 3).GetTypeId() != block.AIR {
			t.Errorf("GetBlockAt(3, %d, 3) = %s, want air", y, w.GetBlockAt(3, y, 3).GetName())
		}
	}
}
