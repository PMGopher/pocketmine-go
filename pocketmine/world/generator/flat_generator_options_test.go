package generator

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	_ "pocketmine-go/pocketmine/world/format/io" // item data handlers, used by the layer parser
)

func TestParseFlatPreset(t *testing.T) {
	options, err := ParseFlatPreset("2;bedrock,59xstone,3xdirt,grass;1;decoration,spawn(radius=10 block=89)")
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		id     int
		height int
	}{{block.BEDROCK, 1}, {block.STONE, 59}, {block.DIRT, 3}, {block.GRASS, 1}}
	if len(options.Structure) != len(want) {
		t.Fatalf("got %d layers, want %d", len(options.Structure), len(want))
	}
	for i, w := range want {
		if l := options.Structure[i]; l.Block.GetTypeId() != w.id || l.Height != w.height {
			t.Errorf("layer %d = %s x%d, want type %d x%d", i, l.Block.GetName(), l.Height, w.id, w.height)
		}
	}
	if options.BiomeID != 1 {
		t.Errorf("biome = %d, want 1", options.BiomeID)
	}
	if _, ok := options.ExtraOptions["decoration"]; !ok {
		t.Error("decoration option missing")
	}
	if got := options.ExtraOptions["spawn"]["radius"]; got != "10" {
		t.Errorf("spawn radius = %q, want 10", got)
	}
	if _, err := ParseFlatPreset("2;notablock;1;"); err == nil {
		t.Error("an unknown layer block must be an invalid preset")
	}
}

// GeneratorManager registers "flat" (a world whose level.dat says "flat" failed to load).
func TestFlatGeneratorIsRegistered(t *testing.T) {
	factory, ok := GetFactory("flat")
	if !ok {
		t.Fatal(`no "flat" generator`)
	}
	gen, err := factory(0, "")
	if err != nil {
		t.Fatal(err)
	}
	chunk := gen.GenerateChunk(0, 0)
	// Default preset "2;bedrock,2xdirt,grass;1;": grass at y=3.
	if got := chunk.GetBlockStateID(0, 3, 0); got != int32(block.VanillaGrass().GetStateId()) {
		t.Errorf("block at y=3 = %d, want grass", got)
	}
	if _, err := factory(0, "2;notablock;1;"); err == nil {
		t.Error("an invalid preset must be rejected")
	}
}

// BedrockWorldData gives vanilla flat worlds the preset "2;7,3,3,2;1" (legacy numeric IDs).
func TestParseFlatPresetWithLegacyIDs(t *testing.T) {
	options, err := ParseFlatPreset("2;7,3,3,2;1")
	if err != nil {
		t.Fatal(err)
	}
	want := []int{block.BEDROCK, block.DIRT, block.DIRT, block.GRASS}
	for i, id := range want {
		if i >= len(options.Structure) || options.Structure[i].Block.GetTypeId() != id {
			t.Fatalf("layers = %v, want bedrock, dirt, dirt, grass", options.Structure)
		}
	}
}
