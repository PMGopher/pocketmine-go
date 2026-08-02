package populator_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/generator/object"
	"pocketmine-go/pocketmine/world/generator/populator"
)

func TestOrePopulatePlacesMaterialInsideReplacesBlocks(t *testing.T) {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	w := world.New(gen, tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
		block.VanillaCoalOre(),
	})
	w.GetOrLoadChunk(0, 0)

	ore := &populator.Ore{}
	ore.SetOreTypes([]*object.OreType{
		object.NewOreType(block.VanillaCoalOre(), block.VanillaStone(), 20, 16, 0, 59),
	})
	ore.Populate(w, 0, 0, utils.NewRandom(1))

	found := false
	for x := 0; x < 16 && !found; x++ {
		for y := 1; y < 59 && !found; y++ {
			for z := 0; z < 16 && !found; z++ {
				if w.GetBlockAt(x, y, z).GetTypeId() == block.COAL_ORE {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected at least one coal ore block to be placed within the stone layer")
	}
}
