package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld() *World {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()))
	return New(gen, tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
	})
}

func TestGetBlockAtReadsGeneratedTerrain(t *testing.T) {
	w := newTestWorld()

	got := w.GetBlockAt(5, 0, 5)
	if got.GetTypeId() != block.BEDROCK {
		t.Errorf("GetBlockAt(5,0,5).GetTypeId() = %d, want BEDROCK (%d)", got.GetTypeId(), block.BEDROCK)
	}

	got = w.GetBlockAt(5, 100, 5)
	if got.GetTypeId() != block.AIR {
		t.Errorf("GetBlockAt(5,100,5).GetTypeId() = %d, want AIR (%d)", got.GetTypeId(), block.AIR)
	}
}

func TestGetBlockAtReturnsAPositionedClone(t *testing.T) {
	w := newTestWorld()

	a := w.GetBlockAt(5, 0, 5)
	b := w.GetBlockAt(5, 0, 5)
	if a == b {
		t.Error("expected two GetBlockAt calls to return distinct instances")
	}
	if pos := a.GetPosition(); pos.FloorX() != 5 || pos.FloorY() != 0 || pos.FloorZ() != 5 {
		t.Errorf("GetPosition() = (%d,%d,%d), want (5,0,5)", pos.FloorX(), pos.FloorY(), pos.FloorZ())
	}
}

func TestSetBlockThenGetBlockAtRoundTrips(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(3, 50, 3, w)

	if err := w.SetBlock(pos, block.VanillaStone()); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}
	got := w.GetBlockAt(3, 50, 3)
	if got.GetTypeId() != block.STONE {
		t.Errorf("GetBlockAt after SetBlock = %d, want STONE (%d)", got.GetTypeId(), block.STONE)
	}
}

func TestSetBlockRegistersUnknownBlockAsATemplate(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(0, 10, 0, w)

	// Grass wasn't in newTestWorld's knownBlocks... it actually is; use a block never registered
	// up front to prove SetBlock's self-registration, not New's.
	glass := block.NewGlass(mustTestBlockIdentifier(block.GLASS), "Test Glass", block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil))
	if err := w.SetBlock(pos, glass); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}

	got := w.GetBlockAt(0, 10, 0)
	if got.GetTypeId() != block.GLASS {
		t.Errorf("GetBlockAt after SetBlock(unregistered type) = %d, want GLASS (%d)", got.GetTypeId(), block.GLASS)
	}
}

func mustTestBlockIdentifier(typeID int) *block.BlockIdentifier {
	idInfo, err := block.NewBlockIdentifier(typeID, nil)
	if err != nil {
		panic(err)
	}
	return idInfo
}

func TestGetOrLoadChunkGeneratesOnFirstAccessAndCaches(t *testing.T) {
	w := newTestWorld()

	first := w.GetOrLoadChunk(0, 0)
	second := w.GetOrLoadChunk(0, 0)
	if first != second {
		t.Error("expected repeated GetOrLoadChunk calls for the same coordinates to return the same chunk")
	}
}

func TestIsInWorldRespectsVerticalBounds(t *testing.T) {
	w := newTestWorld()
	if !w.IsInWorld(0, 0, 0) {
		t.Error("expected Y=0 to be in world")
	}
	if w.IsInWorld(0, YMin-1, 0) {
		t.Error("expected below YMin to be out of world")
	}
	if w.IsInWorld(0, YMax, 0) {
		t.Error("expected YMax itself to be out of world (exclusive upper bound)")
	}
}

func TestUseBreakOnReplacesBlockWithAir(t *testing.T) {
	w := newTestWorld()
	if got := w.GetBlockAt(5, 0, 5); got.GetTypeId() != block.BEDROCK {
		t.Fatalf("precondition failed: expected bedrock at (5,0,5), got %d", got.GetTypeId())
	}

	if !w.UseBreakOn(block.NewPosition(5, 0, 5, w).AsVector3()) {
		t.Fatal("expected UseBreakOn to report success")
	}
	if got := w.GetBlockAt(5, 0, 5); got.GetTypeId() != block.AIR {
		t.Errorf("GetBlockAt after UseBreakOn = %d, want AIR (%d)", got.GetTypeId(), block.AIR)
	}
}

func TestGetOrLoadChunkAtPositionSetsBlockStateID(t *testing.T) {
	w := newTestWorld()
	pos := block.NewPosition(1, 0, 1, w)

	c, ok := w.GetOrLoadChunkAtPosition(pos)
	if !ok {
		t.Fatal("expected GetOrLoadChunkAtPosition to report success")
	}
	c.SetBlockStateID(1, 200, 1, int(block.VanillaStone().GetStateId()))

	if got := w.GetBlockAt(1, 200, 1); got.GetTypeId() != block.STONE {
		t.Errorf("GetTypeId() after chunk adapter Set = %d, want STONE (%d)", got.GetTypeId(), block.STONE)
	}
}
