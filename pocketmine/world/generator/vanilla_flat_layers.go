package generator

import "pocketmine-go/pocketmine/block"

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
