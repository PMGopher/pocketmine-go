// External test package (hell_test, not hell): nether_test.go needs pocketmine-go/pocketmine/world
// for its populator integration test, and world imports pocketmine-go/pocketmine/world/generator,
// which (via its own generator registry) imports this hell package - an internal test file would
// create a real import cycle (hell -> world -> generator -> hell); an external test package avoids
// it, since it's a separate compiled unit from hell itself.
package hell_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/hell"
)

func newTestNetherWorld(t *testing.T) *world.World {
	t.Helper()
	tr := convert.NewBlockTranslator()
	return world.New(hell.NewNether(42), tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaNetherrack(),
		block.VanillaLava(),
		block.VanillaNetherQuartzOre(),
	})
}

func TestGenerateChunkHasBedrockFloorAndCeiling(t *testing.T) {
	chunk := hell.NewNether(42).GenerateChunk(0, 0)

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			if got := chunk.GetBlockStateID(x, 0, z); got != int32(block.VanillaBedrock().GetStateId()) {
				t.Fatalf("GetBlockStateID(%d,0,%d) = %d, want bedrock", x, z, got)
			}
			if got := chunk.GetBlockStateID(x, hell.NetherTerrainHeight-1, z); got != int32(block.VanillaBedrock().GetStateId()) {
				t.Fatalf("GetBlockStateID(%d,%d,%d) = %d, want bedrock", x, hell.NetherTerrainHeight-1, z, got)
			}
		}
	}
}

func TestGenerateChunkFillsTheMiddleWithOnlyAirNetherrackOrLava(t *testing.T) {
	chunk := hell.NewNether(42).GenerateChunk(0, 0)

	air := int32(block.VanillaAir().GetStateId())
	netherrack := int32(block.VanillaNetherrack().GetStateId())
	lava := int32(block.VanillaLava().GetStateId())

	sawNetherrack, sawAirOrLava := false, false
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := 1; y < hell.NetherTerrainHeight-1; y++ {
				switch got := chunk.GetBlockStateID(x, y, z); got {
				case netherrack:
					sawNetherrack = true
				case air, lava:
					sawAirOrLava = true
				default:
					t.Fatalf("GetBlockStateID(%d,%d,%d) = %d, want air, netherrack or lava", x, y, z, got)
				}
			}
		}
	}
	if !sawNetherrack {
		t.Error("no netherrack found anywhere in the chunk - noise-shaping looks broken")
	}
	if !sawAirOrLava {
		t.Error("no air or lava (caves/lava seas) found anywhere in the chunk - noise-shaping looks broken")
	}
}

func TestGenerateChunkUsesUniformHellBiomeAcrossTheEntireVerticalRange(t *testing.T) {
	chunk := hell.NewNether(42).GenerateChunk(0, 0)

	// Includes Y values well outside the 0-127 terrain band (see hell.NetherTerrainHeight) - the
	// biome must still be HELL there too (see GenerateChunk's own doc comment on why every
	// subchunk gets a uniform Hell biome, not just the terrain band).
	for _, y := range []int{-60, 0, 63, 127, 200, 300} {
		if got := chunk.GetBiomeID(5, y, 5); got != int32(biome.IDHell) {
			t.Errorf("GetBiomeID(5,%d,5) = %d, want IDHell (%d)", y, got, biome.IDHell)
		}
	}
}

func TestGenerateChunkIsDeterministicForAFixedSeed(t *testing.T) {
	a := hell.NewNether(42).GenerateChunk(3, -2)
	b := hell.NewNether(42).GenerateChunk(3, -2)

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := 0; y < hell.NetherTerrainHeight; y++ {
				got, want := a.GetBlockStateID(x, y, z), b.GetBlockStateID(x, y, z)
				if got != want {
					t.Fatalf("GetBlockStateID(%d,%d,%d) = %d, want %d (same seed+coords must generate identically)", x, y, z, got, want)
				}
			}
		}
	}
}

func TestGenerateChunkDependsOnChunkCoordinates(t *testing.T) {
	n := hell.NewNether(42)
	a := n.GenerateChunk(0, 0)
	b := n.GenerateChunk(1000, 1000)

	different := false
	for x := 0; x < format.SubChunkEdgeLength && !different; x++ {
		for z := 0; z < format.SubChunkEdgeLength && !different; z++ {
			for y := 1; y < hell.NetherTerrainHeight-1 && !different; y++ {
				if a.GetBlockStateID(x, y, z) != b.GetBlockStateID(x, y, z) {
					different = true
				}
			}
		}
	}
	if !different {
		t.Error("expected chunks (0,0) and (1000,1000) to have at least some different terrain")
	}
}

func TestPopulateChunkPlacesNetherQuartzOreSomewhereInASufficientlyLargeArea(t *testing.T) {
	w := newTestNetherWorld(t)

	quartz := int32(block.VanillaNetherQuartzOre().GetStateId())
	found := false
	for cx := 0; cx < 4 && !found; cx++ {
		for cz := 0; cz < 4 && !found; cz++ {
			chunk := w.GetOrLoadChunk(cx, cz)
			for x := 0; x < format.SubChunkEdgeLength && !found; x++ {
				for z := 0; z < format.SubChunkEdgeLength && !found; z++ {
					for y := 10; y <= 117 && !found; y++ {
						if chunk.GetBlockStateID(x, y, z) == quartz {
							found = true
						}
					}
				}
			}
		}
	}
	if !found {
		t.Error("no nether quartz ore found across a 4x4 chunk area - Ore populator doesn't look wired up")
	}
}
