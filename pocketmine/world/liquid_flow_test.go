package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

// TestWaterSourceFlowsOutward checks Liquid::onScheduledUpdate through the world tick: a water
// source on flat ground spreads 7 blocks in each direction, decaying by 1 per block.
func TestWaterSourceFlowsOutward(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	if err := w.SetBlock(block.NewPosition(8, 64, 8, w), block.VanillaWater()); err != nil {
		t.Fatal(err)
	}
	for tick := 1; tick <= 200; tick++ {
		w.DoTick(int64(tick))
	}
	for dist := 1; dist <= 7; dist++ {
		got, ok := w.GetBlockAt(8+dist, 64, 8).(*block.Water)
		if !ok {
			t.Fatalf("block %d east of the source = %s, want flowing water", dist, w.GetBlockAt(8+dist, 64, 8).GetName())
		}
		if got.Decay != dist || got.Falling {
			t.Errorf("water %d east of the source: decay %d falling %v, want decay %d", dist, got.Decay, got.Falling, dist)
		}
	}
	if got := w.GetBlockAt(16, 64, 8).GetTypeId(); got != block.AIR {
		t.Errorf("8 blocks from the source = %d, want air", got)
	}
}

// TestLavaFlowsTowardsDrop checks MinimumCostFlowCalculator: lava only flows towards the nearest drop.
func TestLavaFlowsTowardsDrop(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	// A hole 2 blocks west of the source: MinimumCostFlowCalculator sends the lava only that way.
	if err := w.SetBlock(block.NewPosition(6, 63, 8, w), block.VanillaAir()); err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(block.NewPosition(8, 64, 8, w), block.VanillaLava()); err != nil {
		t.Fatal(err)
	}
	for tick := 1; tick <= 100; tick++ {
		w.DoTick(int64(tick))
	}
	if _, ok := w.GetBlockAt(7, 64, 8).(*block.Lava); !ok {
		t.Errorf("west of the source = %s, want lava", w.GetBlockAt(7, 64, 8).GetName())
	}
	if _, ok := w.GetBlockAt(9, 64, 8).(*block.Lava); ok {
		t.Error("lava flowed east, away from the drop")
	}
	if _, ok := w.GetBlockAt(6, 63, 8).(*block.Lava); !ok {
		t.Errorf("the hole = %s, want falling lava", w.GetBlockAt(6, 63, 8).GetName())
	}
}
