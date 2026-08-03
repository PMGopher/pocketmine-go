package populator_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/generator/object"
	"pocketmine-go/pocketmine/world/generator/populator"
)

func TestTreePopulatePlacesLogsSomewhereInTheChunk(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	tr := populator.NewTree(object.TreeTypeOak)
	tr.BaseAmount = 20
	tr.Populate(w, 0, 0, utils.NewRandom(3))

	found := false
	for x := 0; x < 16 && !found; x++ {
		for y := 63; y < 80 && !found; y++ {
			for z := 0; z < 16 && !found; z++ {
				if w.GetBlockAt(x, y, z).GetTypeId() == block.OAK_LOG {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected at least one oak log to be placed somewhere in the chunk")
	}
}

func TestTreePopulateWithZeroAmountPlacesNothing(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	tr := populator.NewTree(object.TreeTypeOak)
	tr.BaseAmount = 0
	tr.RandomAmount = 0
	tr.Populate(w, 0, 0, utils.NewRandom(3))

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			if got := w.GetBlockAt(x, 64, z).GetTypeId(); got != block.AIR {
				t.Errorf("GetBlockAt(%d,64,%d) = %d, want AIR (%d) since amount is 0", x, z, got, block.AIR)
			}
		}
	}
}
