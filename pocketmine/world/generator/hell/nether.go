// Package hell is a port of pocketmine\world\generator\hell: just Nether, the Nether dimension's
// generator (named after its own PHP namespace, one level up from the more usual generator/normal
// split).
package hell

import (
	"math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/noise"
	"pocketmine-go/pocketmine/world/generator/object"
	"pocketmine-go/pocketmine/world/generator/populator"
)

// NetherTerrainHeight mirrors the hardcoded 128 Nether::generateChunk's own terrain loop uses -
// Nether terrain has always occupied a fixed 0-127 band regardless of the overworld's own Y_MIN/
// Y_MAX range (real PHP hardcodes `$y < 128` directly, not derived from World::Y_MAX), so this is
// a bare literal matching that, not derived from anything else.
const NetherTerrainHeight = 128

// Nether is a port of pocketmine\world\generator\hell\Nether.
//
// One documented simplification: real PHP looks up the biome to populate via
// `BiomeRegistry::getInstance()->getBiome($chunk->getBiomeId(7,7,7))` - a generic "whatever biome
// this column ended up as" lookup. Nether's biome is uniformly HELL everywhere (no per-column
// variation the way Normal's pickBiome has), so that lookup can only ever resolve to one biome -
// this just holds that one *biome.Biome directly instead of a whole biome.Registry to look it up
// in every call, an equivalent result reached more directly given Nether's own fixed biome.
type Nether struct {
	seed int

	waterHeight    int
	emptyHeight    int
	emptyAmplitude float64
	density        float64

	random    *utils.Random
	noiseBase *noise.Simplex
	hellBiome *biome.Biome

	populators []populator.Populator

	airStateID        int32
	bedrockStateID    int32
	netherrackStateID int32
	stillLavaStateID  int32
}

// NewNether is a port of Nether::__construct.
func NewNether(seed int) *Nether {
	n := &Nether{
		seed:           seed,
		waterHeight:    32,
		emptyHeight:    64,
		emptyAmplitude: 1,
		density:        0.5,
		hellBiome:      biome.NewHellBiome(),
	}

	random := utils.NewRandom(seed)
	n.noiseBase = noise.NewSimplex(random, 4, 1.0/4, 1.0/64)
	random.SetSeed(seed)
	n.random = random

	ore := &populator.Ore{}
	ore.SetOreTypes([]*object.OreType{
		object.NewOreType(block.VanillaNetherQuartzOre(), block.VanillaNetherrack(), 16, 14, 10, 117),
	})
	n.populators = []populator.Populator{ore}

	n.airStateID = int32(block.VanillaAir().GetStateId())
	n.bedrockStateID = int32(block.VanillaBedrock().GetStateId())
	n.netherrackStateID = int32(block.VanillaNetherrack().GetStateId())
	n.stillLavaStateID = int32(block.VanillaLava().GetStateId())

	return n
}

// GenerateChunk is a port of Nether::generateChunk. Every subchunk's biome comes out uniformly
// HELL via format.NewChunk's own defaultBiomeID auto-fill (see NewChunk's doc comment) rather than
// an explicit per-position SetBiomeID loop across the full Y_MIN..Y_MAX range like the PHP
// original - the result is identical (every position in this chunk is biome HELL either way), just
// reached through the format package's existing default-fill mechanism instead of 3D-looping over
// it by hand.
func (n *Nether) GenerateChunk(chunkX, chunkZ int) *format.Chunk {
	n.random.SetSeed(0xdeadbeef ^ (chunkX << 8) ^ chunkZ ^ n.seed)

	noiseCube := n.noiseBase.GetFastNoise3D(
		format.SubChunkEdgeLength, NetherTerrainHeight, format.SubChunkEdgeLength,
		4, 8, 4,
		chunkX*format.SubChunkEdgeLength, 0, chunkZ*format.SubChunkEdgeLength,
	)

	chunk := format.NewChunk(nil, false, n.airStateID, int32(biome.IDHell))

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := 0; y < NetherTerrainHeight; y++ {
				if y == 0 || y == NetherTerrainHeight-1 {
					chunk.SetBlockStateID(x, y, z, n.bedrockStateID)
					continue
				}

				noiseValue := (math.Abs(float64(n.emptyHeight-y))/float64(n.emptyHeight))*n.emptyAmplitude - noiseCube[x][z][y]
				noiseValue -= 1 - n.density

				if noiseValue > 0 {
					chunk.SetBlockStateID(x, y, z, n.netherrackStateID)
				} else if y <= n.waterHeight {
					chunk.SetBlockStateID(x, y, z, n.stillLavaStateID)
				}
			}
		}
	}

	return chunk
}

// PopulateChunk is a port of Nether::populateChunk - see Nether's own doc comment on why this
// calls hellBiome.PopulateChunk directly instead of looking it up via a biome.Registry.
func (n *Nether) PopulateChunk(world block.World, chunkX, chunkZ int) {
	n.random.SetSeed(0xdeadbeef ^ (chunkX << 8) ^ chunkZ ^ n.seed)
	for _, p := range n.populators {
		p.Populate(world, chunkX, chunkZ, n.random)
	}

	n.hellBiome.PopulateChunk(world, chunkX, chunkZ, n.random)
}
