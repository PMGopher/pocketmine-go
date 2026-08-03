package populator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/object"
)

// Tree is a port of pocketmine\world\generator\populator\Tree.
type Tree struct {
	RandomAmount int
	BaseAmount   int
	Type         object.TreeType
}

// NewTree matches the PHP constructor's implicit defaults (randomAmount=1, baseAmount=0) plus its
// `?TreeType $type = null` defaulting to oak.
func NewTree(t object.TreeType) *Tree {
	return &Tree{RandomAmount: 1, BaseAmount: 0, Type: t}
}

func (t *Tree) Populate(world block.World, chunkX, chunkZ int, random *utils.Random) {
	amount := random.NextRange(0, t.RandomAmount) + t.BaseAmount

	for i := 0; i < amount; i++ {
		// Matches Tree.php's own random.nextRange bounds exactly (inclusive of
		// chunkX*16+EDGE_LENGTH, one past the local chunk edge) - not TallGrass's EDGE_LENGTH-1
		// bound. An occasional tree seed landing in the neighbouring chunk this way is real
		// upstream behaviour, not a typo to "fix".
		x := random.NextRange(chunkX<<4, (chunkX<<4)+format.SubChunkEdgeLength)
		z := random.NextRange(chunkZ<<4, (chunkZ<<4)+format.SubChunkEdgeLength)

		y, ok := getHighestWorkableBlockForTree(world, x, z)
		if !ok {
			continue
		}

		tree := object.NewTreeFromType(random, t.Type)
		if tx := tree.GetBlockTransaction(world, x, y, z, random); tx != nil {
			tx.Apply()
		}
	}
}

// typeTagged lets getHighestWorkableBlockForTree reach the promoted-but-not-exported-on-Behavior
// HasTypeTag method - same forward-compatible local-interface pattern as this port's
// positionable/asItemOrNil.
type typeTagged interface{ HasTypeTag(tag string) bool }

func hasTypeTag(blk block.Behavior, tag string) bool {
	tt, ok := blk.(typeTagged)
	return ok && tt.HasTypeTag(tag)
}

// getHighestWorkableBlockForTree is a port of populator\Tree's own getHighestWorkableBlock -
// distinct from TallGrass's (this one stops on dirt/mud, not "not air/leaves/snow").
func getHighestWorkableBlockForTree(world block.World, x, z int) (int, bool) {
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
		if hasTypeTag(b, block.BlockTypeTagsDirt) || hasTypeTag(b, block.BlockTypeTagsMud) {
			return y + 1, true
		}
		if b.GetTypeId() != block.AIR && b.GetTypeId() != block.SNOW_LAYER {
			return -1, false
		}
	}

	return -1, false
}
