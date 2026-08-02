package generator

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

func newTestEmptyStateID() int32 {
	return int32(block.VanillaAir().GetStateId())
}

func TestNewFlatBuildsLayersInOrder(t *testing.T) {
	layers := []FlatLayer{
		{Block: block.VanillaBedrock(), Height: 1},
		{Block: block.VanillaStone(), Height: 2},
	}
	f := NewFlat(layers, 1, newTestEmptyStateID())
	chunk := f.GenerateChunk(0, 0)

	bedrockID := int32(block.VanillaBedrock().GetStateId())
	stoneID := int32(block.VanillaStone().GetStateId())
	airID := newTestEmptyStateID()

	if got := chunk.GetBlockStateID(5, 0, 5); got != bedrockID {
		t.Errorf("Y=0 = %d, want bedrock (%d)", got, bedrockID)
	}
	if got := chunk.GetBlockStateID(5, 1, 5); got != stoneID {
		t.Errorf("Y=1 = %d, want stone (%d)", got, stoneID)
	}
	if got := chunk.GetBlockStateID(5, 2, 5); got != stoneID {
		t.Errorf("Y=2 = %d, want stone (%d)", got, stoneID)
	}
	if got := chunk.GetBlockStateID(5, 3, 5); got != airID {
		t.Errorf("Y=3 (above the structure) = %d, want air (%d)", got, airID)
	}
}

func TestNewFlatFillsEveryColumnInEverySubChunk(t *testing.T) {
	f := NewFlat(VanillaFlatLayers(), VanillaFlatBiomeID, newTestEmptyStateID())
	chunk := f.GenerateChunk(0, 0)

	bedrockID := int32(block.VanillaBedrock().GetStateId())
	for _, pos := range [][2]int{{0, 0}, {15, 0}, {0, 15}, {15, 15}, {7, 7}} {
		if got := chunk.GetBlockStateID(pos[0], 0, pos[1]); got != bedrockID {
			t.Errorf("GetBlockStateID(%d,0,%d) = %d, want bedrock (%d)", pos[0], pos[1], got, bedrockID)
		}
	}
}

func TestGenerateChunkReturnsIndependentClones(t *testing.T) {
	f := NewFlat(VanillaFlatLayers(), VanillaFlatBiomeID, newTestEmptyStateID())

	a := f.GenerateChunk(0, 0)
	b := f.GenerateChunk(1, 1)

	a.SetBlockStateID(0, 0, 0, 999)
	if got := b.GetBlockStateID(0, 0, 0); got == 999 {
		t.Error("expected chunks from separate GenerateChunk calls to be independent")
	}
}

func TestVanillaFlatLayersMatchesTheClassicPreset(t *testing.T) {
	layers := VanillaFlatLayers()
	totalHeight := 0
	for _, l := range layers {
		totalHeight += l.Height
	}
	// bedrock(1) + stone(59) + dirt(3) + grass(1) = 64, the classic "2;bedrock,59xstone,3xdirt,grass;1;" preset.
	if totalHeight != 64 {
		t.Errorf("total layer height = %d, want 64", totalHeight)
	}
}
