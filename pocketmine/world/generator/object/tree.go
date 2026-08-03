package object

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
)

// Tree is a port of pocketmine\world\generator\object\Tree, folded together with its 3 concrete
// subclasses this port needs (OakTree/SpruceTree/BirchTree - see NewOakTree/NewSpruceTree/
// NewBirchTree). PHP models the two points of variation (how tall a rolled tree ends up, and how
// the canopy is shaped) as overridden methods on subclasses; since only SpruceTree overrides
// anything beyond the height roll, this port uses plain function fields instead of a deeper type
// hierarchy - the same "function value instead of a one-off abstract subtype" pattern already used
// elsewhere in this port (e.g. biomeselector.LookupFunc). JungleTree/AcaciaTree/AzaleaTree/
// NetherTree aren't ported - no biome this port registers uses them (see TreeFactory's doc
// comment).
type Tree struct {
	TrunkBlock block.Behavior
	LeafBlock  block.Behavior
	TreeHeight int

	// rollHeight mirrors each subclass's own getBlockTransaction override, which recomputes
	// TreeHeight from a fresh random roll right before generating (e.g. OakTree's
	// `$this->treeHeight = $random->nextBoundedInt(3) + 4`). Never nil in practice - every
	// constructor here sets one.
	rollHeight func(random *utils.Random) int

	// generateTrunkHeight overrides Tree::generateTrunkHeight - only SpruceTree does; nil means use
	// the base class's `treeHeight - 1`.
	generateTrunkHeight func(treeHeight int, random *utils.Random) int

	// placeCanopy overrides Tree::placeCanopy - only SpruceTree does (a layered-ring shape instead
	// of the base class's rounded blob); nil means use defaultPlaceCanopy.
	placeCanopy func(t *Tree, x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl)
}

// CanPlaceObject is a port of Tree::canPlaceObject.
func (t *Tree) CanPlaceObject(world block.World, x, y, z int, random *utils.Random) bool {
	radiusToCheck := 0
	for yy := 0; yy < t.TreeHeight+3; yy++ {
		if yy == 1 || yy == t.TreeHeight {
			radiusToCheck++
		}
		for xx := -radiusToCheck; xx < radiusToCheck+1; xx++ {
			for zz := -radiusToCheck; zz < radiusToCheck+1; zz++ {
				if !canOverride(world.GetBlockAt(x+xx, y+yy, z+zz)) {
					return false
				}
			}
		}
	}
	return true
}

// GetBlockTransaction is a port of Tree::getBlockTransaction, plus each subclass's own override
// that rerolls TreeHeight first (via rollHeight) before doing anything else - matching e.g.
// OakTree::getBlockTransaction setting $this->treeHeight before calling parent::.
func (t *Tree) GetBlockTransaction(world block.World, x, y, z int, random *utils.Random) *block.BlockTransactionImpl {
	if t.rollHeight != nil {
		t.TreeHeight = t.rollHeight(random)
	}
	if !t.CanPlaceObject(world, x, y, z, random) {
		return nil
	}

	tx := block.NewBlockTransaction(world)
	t.placeTrunk(x, y, z, random, t.trunkHeight(random), tx)
	if t.placeCanopy != nil {
		t.placeCanopy(t, x, y, z, random, tx)
	} else {
		t.defaultPlaceCanopy(x, y, z, random, tx)
	}
	return tx
}

func (t *Tree) trunkHeight(random *utils.Random) int {
	if t.generateTrunkHeight != nil {
		return t.generateTrunkHeight(t.TreeHeight, random)
	}
	return t.TreeHeight - 1
}

// placeTrunk is a port of Tree::placeTrunk.
func (t *Tree) placeTrunk(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl) {
	tx.AddBlockAt(x, y-1, z, block.VanillaDirt())

	for yy := 0; yy < trunkHeight; yy++ {
		if canOverride(tx.FetchBlockAt(x, y+yy, z)) {
			tx.AddBlockAt(x, y+yy, z, t.TrunkBlock.Clone())
		}
	}
}

// defaultPlaceCanopy is a port of Tree::placeCanopy (the base class's rounded-blob shape).
func (t *Tree) defaultPlaceCanopy(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
	for yy := y - 3 + t.TreeHeight; yy <= y+t.TreeHeight; yy++ {
		yOff := yy - (y + t.TreeHeight)
		mid := int(1 - float64(yOff)/2)
		for xx := x - mid; xx <= x+mid; xx++ {
			xOff := absInt(xx - x)
			for zz := z - mid; zz <= z+mid; zz++ {
				zOff := absInt(zz - z)
				if xOff == mid && zOff == mid && (yOff == 0 || random.NextBoundedInt(2) == 0) {
					continue
				}
				if !tx.FetchBlockAt(xx, yy, zz).IsSolid() {
					tx.AddBlockAt(xx, yy, zz, t.LeafBlock.Clone())
				}
			}
		}
	}
}

// canOverride is a port of Tree::canOverride.
func canOverride(blk block.Behavior) bool {
	if blk.CanBeReplaced() {
		return true
	}
	if _, ok := blk.(*block.Sapling); ok {
		return true
	}
	if _, ok := blk.(*block.Leaves); ok {
		return true
	}
	return false
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// NewOakTree is a port of OakTree.
func NewOakTree() *Tree {
	return &Tree{
		TrunkBlock: block.VanillaOakLog(),
		LeafBlock:  block.VanillaOakLeaves(),
		TreeHeight: 7,
		rollHeight: func(random *utils.Random) int { return random.NextBoundedInt(3) + 4 },
	}
}

// NewBirchTree is a port of BirchTree. superBirch matches its $superBirch constructor param (a
// taller variant TreeFactory rolls a 1/39 chance of).
func NewBirchTree(superBirch bool) *Tree {
	return &Tree{
		TrunkBlock: block.VanillaBirchLog(),
		LeafBlock:  block.VanillaBirchLeaves(),
		TreeHeight: 7,
		rollHeight: func(random *utils.Random) int {
			h := random.NextBoundedInt(3) + 5
			if superBirch {
				h += 5
			}
			return h
		},
	}
}

// NewSpruceTree is a port of SpruceTree.
func NewSpruceTree() *Tree {
	return &Tree{
		TrunkBlock: block.VanillaSpruceLog(),
		LeafBlock:  block.VanillaSpruceLeaves(),
		TreeHeight: 10,
		rollHeight: func(random *utils.Random) int { return random.NextBoundedInt(4) + 6 },
		generateTrunkHeight: func(treeHeight int, random *utils.Random) int {
			return treeHeight - random.NextBoundedInt(3)
		},
		placeCanopy: spruceCanopy,
	}
}

// spruceCanopy is a port of SpruceTree::placeCanopy (layered rings shrinking then holding steady,
// instead of the base class's single rounded blob).
func spruceCanopy(t *Tree, x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
	topSize := t.TreeHeight - (1 + random.NextBoundedInt(2))
	lRadius := 2 + random.NextBoundedInt(2)
	radius := random.NextBoundedInt(2)
	maxR := 1
	minR := 0

	for yy := 0; yy <= topSize; yy++ {
		yyy := y + t.TreeHeight - yy

		for xx := x - radius; xx <= x+radius; xx++ {
			xOff := absInt(xx - x)
			for zz := z - radius; zz <= z+radius; zz++ {
				zOff := absInt(zz - z)
				if xOff == radius && zOff == radius && radius > 0 {
					continue
				}
				if !tx.FetchBlockAt(xx, yyy, zz).IsSolid() {
					tx.AddBlockAt(xx, yyy, zz, t.LeafBlock.Clone())
				}
			}
		}

		if radius >= maxR {
			radius = minR
			minR = 1
			maxR++
			if maxR > lRadius {
				maxR = lRadius
			}
		} else {
			radius++
		}
	}
}
