package generator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/generator/object"
	"pocketmine-go/pocketmine/world/generator/populator"
)

// VanillaFlatLayers builds the classic PocketMine-MP default flat preset
// ("2;bedrock,59xstone,3xdirt,grass;1;", the commented example in resources/pocketmine.yml) as
// []FlatLayer directly, since NewFlat doesn't parse preset strings - see Flat's doc comment.
func VanillaFlatLayers() []FlatLayer {
	return []FlatLayer{
		{Block: block.VanillaBedrock(), Height: 1},
		{Block: block.VanillaStone(), Height: 59},
		{Block: block.VanillaDirt(), Height: 3},
		{Block: block.VanillaGrass(), Height: 1},
	}
}

// VanillaFlatBiomeID is the biome ID used by the classic default flat preset ("1" in
// "2;bedrock,59xstone,3xdirt,grass;1;") - Plains, matching Bedrock's well-known legacy numeric
// biome ID scheme. pocketmine\data\bedrock\BiomeIds isn't vendored in this port (see
// ChunkSerializer's doc comment on the same gap), so this is a bare literal rather than a named
// constant reference.
const VanillaFlatBiomeID int32 = 1

// VanillaFlatOreTypes builds the []*object.OreType list Flat's real constructor sets up when the
// preset's "decoration" extra option is present - see Flat.php's constructor for the exact
// (material, replaces, clusterCount, clusterSize, minHeight, maxHeight) tuples, copied verbatim.
func VanillaFlatOreTypes() []*object.OreType {
	stone := block.VanillaStone()
	return []*object.OreType{
		object.NewOreType(block.VanillaCoalOre(), stone, 20, 16, 0, 128),
		object.NewOreType(block.VanillaIronOre(), stone, 20, 8, 0, 64),
		object.NewOreType(block.VanillaRedstoneOre(), stone, 8, 7, 0, 16),
		object.NewOreType(block.VanillaLapisLazuliOre(), stone, 1, 6, 0, 32),
		object.NewOreType(block.VanillaGoldOre(), stone, 2, 8, 0, 32),
		object.NewOreType(block.VanillaDiamondOre(), stone, 1, 7, 0, 16),
		object.NewOreType(block.VanillaDirt(), stone, 20, 32, 0, 128),
		object.NewOreType(block.VanillaGravel(), stone, 10, 16, 0, 128),
	}
}

// VanillaFlatDecorationPopulators builds the []populator.Populator list Flat's real constructor
// sets up when the preset's "decoration" extra option is present: a single Ore populator carrying
// VanillaFlatOreTypes.
func VanillaFlatDecorationPopulators() []populator.Populator {
	ore := &populator.Ore{}
	ore.SetOreTypes(VanillaFlatOreTypes())
	return []populator.Populator{ore}
}
