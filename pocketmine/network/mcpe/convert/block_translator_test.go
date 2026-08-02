package convert

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

func newTestAir() *block.Air {
	idInfo, err := block.NewBlockIdentifier(block.AIR, nil)
	if err != nil {
		panic(err)
	}
	a := block.NewAir(idInfo, "Air", block.NewBlockTypeInfo(block.BlockBreakInfoIndestructible(-1.0), nil, nil))
	return a
}

func newTestStone() *block.Stone {
	idInfo, err := block.NewBlockIdentifier(block.STONE, nil)
	if err != nil {
		panic(err)
	}
	return block.NewStone(idInfo, "Stone", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))
}

func newTestBedrock() *block.Bedrock {
	idInfo, err := block.NewBlockIdentifier(block.BEDROCK, nil)
	if err != nil {
		panic(err)
	}
	return block.NewBedrock(idInfo, "Bedrock", block.NewBlockTypeInfo(block.BlockBreakInfoIndestructible(18000000), nil, nil))
}

func TestBlockTranslatorTranslatesRegisteredBlocks(t *testing.T) {
	tr := NewBlockTranslator()

	if id := tr.InternalIDToNetworkID(newTestAir()); id != 13094 {
		t.Errorf("air runtime ID = %d, want 13094", id)
	}
	if id := tr.InternalIDToNetworkID(newTestStone()); id != 2706 {
		t.Errorf("stone runtime ID = %d, want 2706", id)
	}
}

func TestBlockTranslatorTranslatesBedrockWithState(t *testing.T) {
	tr := NewBlockTranslator()

	b := newTestBedrock()
	b.BurnsForeverFlag = false
	if id := tr.InternalIDToNetworkID(b); id != 13805 {
		t.Errorf("bedrock(burns=false) runtime ID = %d, want 13805", id)
	}
}

func TestBlockTranslatorCachesByInternalStateID(t *testing.T) {
	tr := NewBlockTranslator()

	first := tr.InternalIDToNetworkID(newTestStone())
	second := tr.InternalIDToNetworkID(newTestStone())
	if first != second {
		t.Errorf("expected repeated calls for the same internal state to return the same ID, got %d then %d", first, second)
	}
	if len(tr.networkIDCache) != 1 {
		t.Errorf("expected exactly 1 cache entry, got %d", len(tr.networkIDCache))
	}
}

func TestBlockTranslatorFallsBackForUnregisteredBlockType(t *testing.T) {
	tr := NewBlockTranslator()

	// GRASS_PATH has no registered BlockStateSerializer yet.
	idInfo, err := block.NewBlockIdentifier(block.GRASS_PATH, nil)
	if err != nil {
		t.Fatal(err)
	}
	unregistered := block.NewGrassPath(idInfo, "Grass Path", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))

	got := tr.InternalIDToNetworkID(unregistered)
	if got != tr.FallbackStateID() {
		t.Errorf("expected the fallback state ID (%d) for an unregistered block, got %d", tr.FallbackStateID(), got)
	}
}
