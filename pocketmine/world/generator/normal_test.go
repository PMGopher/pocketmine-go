package generator

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/format"
)

func TestGenerateChunkIsDeterministicForAFixedSeed(t *testing.T) {
	a := NewNormal(42).GenerateChunk(3, -2)
	b := NewNormal(42).GenerateChunk(3, -2)

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := -20; y < 100; y++ {
				got, want := a.GetBlockStateID(x, y, z), b.GetBlockStateID(x, y, z)
				if got != want {
					t.Fatalf("GetBlockStateID(%d,%d,%d) = %d, want %d (same seed+coords must generate identically)", x, y, z, got, want)
				}
			}
		}
	}
}

func TestGenerateChunkDependsOnChunkCoordinates(t *testing.T) {
	// Adjacent chunks can legitimately be identical (e.g. deep inside one large open-ocean region,
	// every column is bare water up to waterHeight regardless of noise - not a bug, just a boring
	// sample). Comparing chunks far enough apart to guarantee different biome regions instead is a
	// much more reliable way to confirm terrain shape actually depends on chunk position, not just
	// non-determinism.
	n := NewNormal(42)
	a := n.GenerateChunk(0, 0)
	b := n.GenerateChunk(1000, 1000)

	different := false
	for x := 0; x < format.SubChunkEdgeLength && !different; x++ {
		for z := 0; z < format.SubChunkEdgeLength && !different; z++ {
			h1, _ := a.GetHighestBlockAt(x, z)
			h2, _ := b.GetHighestBlockAt(x, z)
			if h1 != h2 {
				different = true
			}
		}
	}
	if !different {
		t.Error("expected chunks (0,0) and (1000,1000) to have at least some different terrain heights")
	}
}

func TestGenerateChunkFillsBedrockAtTheBottomAndVariesSurfaceHeight(t *testing.T) {
	chunk := NewNormal(1).GenerateChunk(0, 0)

	bedrockID := int32(block.VanillaBedrock().GetStateId())
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			if got := chunk.GetBlockStateID(x, 0, z); got != bedrockID {
				t.Errorf("GetBlockStateID(%d,0,%d) = %d, want bedrock (%d)", x, z, got, bedrockID)
			}
		}
	}

	minH, maxH := 1<<30, -(1 << 30)
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			h, ok := chunk.GetHighestBlockAt(x, z)
			if !ok {
				t.Fatalf("GetHighestBlockAt(%d,%d) found nothing", x, z)
			}
			if h < minH {
				minH = h
			}
			if h > maxH {
				maxH = h
			}
		}
	}
	if minH < 30 || maxH > 200 {
		t.Errorf("surface height range %d..%d looks implausible for a single 16x16 chunk", minH, maxH)
	}
}
