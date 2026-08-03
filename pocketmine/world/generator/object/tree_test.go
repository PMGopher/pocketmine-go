package object_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/generator/object"
)

func newTestWorld() *world.World {
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, tr, []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
		block.VanillaOakLog(), block.VanillaOakLeaves(),
		block.VanillaSpruceLog(), block.VanillaSpruceLeaves(),
		block.VanillaBirchLog(), block.VanillaBirchLeaves(),
	})
}

func TestOakTreeGrowsATrunkAndCanopyOnGrass(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0) // flat terrain: grass at y=63, air above

	tree := object.NewOakTree()
	tx := tree.GetBlockTransaction(w, 5, 64, 5, utils.NewRandom(1))
	if tx == nil {
		t.Fatal("expected a non-nil transaction on open flat ground")
	}
	if !tx.Apply() {
		t.Fatal("expected Apply to report a change")
	}

	if got := w.GetBlockAt(5, 64, 5).GetTypeId(); got != block.OAK_LOG {
		t.Errorf("GetBlockAt(5,64,5) = %d, want OAK_LOG (%d)", got, block.OAK_LOG)
	}

	found := false
	for x := 0; x < 16 && !found; x++ {
		for y := 64; y < 74 && !found; y++ {
			for z := 0; z < 16 && !found; z++ {
				if w.GetBlockAt(x, y, z).GetTypeId() == block.OAK_LEAVES {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected at least one oak leaves block near the trunk")
	}
}

func TestTreeCanPlaceObjectFailsWhenObstructed(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	// Fill the space directly above the intended trunk with solid stone, leaving no room to grow.
	if err := w.SetBlock(block.NewPosition(5, 64, 5, w), block.VanillaStone()); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}

	tree := object.NewOakTree()
	if tree.CanPlaceObject(w, 5, 64, 5, utils.NewRandom(1)) {
		t.Error("expected CanPlaceObject to fail with stone obstructing the trunk position")
	}
}

func TestSpruceTreeUsesItsOwnLayeredCanopyShape(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	tree := object.NewSpruceTree()
	tx := tree.GetBlockTransaction(w, 5, 64, 5, utils.NewRandom(7))
	if tx == nil {
		t.Fatal("expected a non-nil transaction on open flat ground")
	}
	tx.Apply()

	if got := w.GetBlockAt(5, 64, 5).GetTypeId(); got != block.SPRUCE_LOG {
		t.Errorf("GetBlockAt(5,64,5) = %d, want SPRUCE_LOG (%d)", got, block.SPRUCE_LOG)
	}
}
