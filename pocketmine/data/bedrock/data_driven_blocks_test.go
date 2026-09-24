package bedrock

import "testing"

func TestDataDrivenBlocksAreInTheBlockPalette(t *testing.T) {
	blocks := DataDrivenBlocks()
	if len(blocks) == 0 {
		t.Fatal("no data-driven blocks loaded")
	}
	for _, b := range blocks {
		if len(b.Components) == 0 {
			t.Errorf("%s has no components", b.Name)
		}
		loadBlockStates()
		if len(blockStatesByName[b.Name]) == 0 {
			t.Errorf("%s isn't in canonical_block_states.nbt", b.Name)
		}
	}
}
