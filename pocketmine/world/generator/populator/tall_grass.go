package populator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
)

// TallGrass is a port of pocketmine\world\generator\populator\TallGrass: scatters
// block.VanillaTallGrass() on top of grass blocks within the chunk.
type TallGrass struct {
	// RandomAmount/BaseAmount are ports of the private $randomAmount/$baseAmount fields (set via
	// setRandomAmount/setBaseAmount in PHP) - exported directly here since this port has no
	// equivalent need to hide them behind setters.
	RandomAmount int
	BaseAmount   int
}

// NewTallGrass matches the PHP constructor's implicit defaults (randomAmount=1, baseAmount=0).
func NewTallGrass() *TallGrass {
	return &TallGrass{RandomAmount: 1, BaseAmount: 0}
}

func (t *TallGrass) Populate(world block.World, chunkX, chunkZ int, random *utils.Random) {
	amount := random.NextRange(0, t.RandomAmount) + t.BaseAmount

	for i := 0; i < amount; i++ {
		x := random.NextRange(chunkX*format.SubChunkEdgeLength, chunkX*format.SubChunkEdgeLength+(format.SubChunkEdgeLength-1))
		z := random.NextRange(chunkZ*format.SubChunkEdgeLength, chunkZ*format.SubChunkEdgeLength+(format.SubChunkEdgeLength-1))
		y, ok := getHighestWorkableBlock(world, x, z)

		if ok && canTallGrassStay(world, x, y, z) {
			_ = world.SetBlock(block.NewPosition(float64(x), float64(y), float64(z), world), block.VanillaTallGrass())
		}
	}
}

func canTallGrassStay(world block.World, x, y, z int) bool {
	b := world.GetBlockAt(x, y, z).GetTypeId()
	return (b == block.AIR || b == block.SNOW_LAYER) && world.GetBlockAt(x, y-1, z).GetTypeId() == block.GRASS
}

// getHighestWorkableBlock is a port of TallGrass::getHighestWorkableBlock.
func getHighestWorkableBlock(world block.World, x, z int) (int, bool) {
	pos := block.NewPosition(float64(x), 0, float64(z), world)
	chunk, ok := world.GetOrLoadChunkAtPosition(pos)
	if !ok {
		return -1, false
	}
	highestBlock, ok := chunk.GetHighestBlockAt(x&(format.SubChunkEdgeLength-1), z&(format.SubChunkEdgeLength-1))
	if !ok {
		return -1, false
	}

	for y := highestBlock; y >= 0; y-- {
		b := world.GetBlockAt(x, y, z)
		if _, isLeaves := b.(*block.Leaves); b.GetTypeId() != block.AIR && !isLeaves && b.GetTypeId() != block.SNOW_LAYER {
			return y + 1, true
		}
	}

	return -1, false
}
