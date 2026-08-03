package populator_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/generator/populator"
)

func newTestWorld() *world.World {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
		block.VanillaTallGrass(),
		block.VanillaOakLog(),
		block.VanillaOakLeaves(),
	})
}

func TestTallGrassPopulatePlacesGrassOnTopOfGrassBlocks(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	tg := populator.NewTallGrass()
	tg.BaseAmount = 50
	tg.Populate(w, 0, 0, utils.NewRandom(42))

	found := false
	for x := 0; x < 16 && !found; x++ {
		for z := 0; z < 16 && !found; z++ {
			if w.GetBlockAt(x, 64, z).GetTypeId() == block.TALL_GRASS {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected at least one tall grass block to be placed at y=64 across the chunk")
	}
}

func TestTallGrassPopulateWithZeroAmountPlacesNothing(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	tg := populator.NewTallGrass()
	tg.BaseAmount = 0
	tg.RandomAmount = 0
	tg.Populate(w, 0, 0, utils.NewRandom(42))

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			if got := w.GetBlockAt(x, 64, z).GetTypeId(); got != block.AIR {
				t.Errorf("GetBlockAt(%d,64,%d) = %d, want AIR (%d) since amount is 0", x, z, got, block.AIR)
			}
		}
	}
}
