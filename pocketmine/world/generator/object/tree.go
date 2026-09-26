package object

import (
	"math/rand"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
)

// Tree is a port of pocketmine\world\generator\object\Tree and its subclasses (OakTree,
// SpruceTree, BirchTree, JungleTree, AcaciaTree, AzaleaTree, NetherTree). PHP models each point of
// variation as an overridden method; here each is a function field (nil means the base Tree
// behaviour), set by the constructors below.
type Tree struct {
	TrunkBlock block.Behavior
	LeafBlock  block.Behavior
	TreeHeight int

	// beforeTransaction is the subclass's getBlockTransaction override, which runs before the base
	// one (rerolling the height, resetting per-growth state).
	beforeTransaction func(random *utils.Random)
	// canPlaceObjectFn overrides Tree::canPlaceObject.
	canPlaceObjectFn func(world block.World, x, y, z int, random *utils.Random) bool
	// generateTrunkHeight overrides Tree::generateTrunkHeight.
	generateTrunkHeight func(random *utils.Random) int
	// placeTrunkFn overrides Tree::placeTrunk.
	placeTrunkFn func(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl)
	// placeCanopyFn overrides Tree::placeCanopy.
	placeCanopyFn func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl)
	// canOverrideFn overrides Tree::canOverride.
	canOverrideFn func(blk block.Behavior) bool
}

func newTree(trunkBlock, leafBlock block.Behavior, treeHeight int) *Tree {
	return &Tree{TrunkBlock: trunkBlock, LeafBlock: leafBlock, TreeHeight: treeHeight}
}

// CanPlaceObject is a port of Tree::canPlaceObject.
func (t *Tree) CanPlaceObject(world block.World, x, y, z int, random *utils.Random) bool {
	if t.canPlaceObjectFn != nil {
		return t.canPlaceObjectFn(world, x, y, z, random)
	}
	radiusToCheck := 0
	for yy := 0; yy < t.TreeHeight+3; yy++ {
		if yy == 1 || yy == t.TreeHeight {
			radiusToCheck++
		}
		for xx := -radiusToCheck; xx < radiusToCheck+1; xx++ {
			for zz := -radiusToCheck; zz < radiusToCheck+1; zz++ {
				if !t.canOverride(world.GetBlockAt(x+xx, y+yy, z+zz)) {
					return false
				}
			}
		}
	}
	return true
}

// GetBlockTransaction is a port of Tree::getBlockTransaction: the blocks the tree would change
// when growing at x,y,z, or nil if it can't grow there.
func (t *Tree) GetBlockTransaction(world block.World, x, y, z int, random *utils.Random) *block.BlockTransactionImpl {
	if t.beforeTransaction != nil {
		t.beforeTransaction(random)
	}
	if !t.CanPlaceObject(world, x, y, z, random) {
		return nil
	}

	tx := block.NewBlockTransaction(world)
	trunkHeight := t.TreeHeight - 1
	if t.generateTrunkHeight != nil {
		trunkHeight = t.generateTrunkHeight(random)
	}
	if t.placeTrunkFn != nil {
		t.placeTrunkFn(x, y, z, random, trunkHeight, tx)
	} else {
		t.placeTrunk(x, y, z, random, trunkHeight, tx)
	}
	if t.placeCanopyFn != nil {
		t.placeCanopyFn(x, y, z, random, tx)
	} else {
		t.placeCanopy(x, y, z, random, tx)
	}
	return tx
}

// placeTrunk is a port of Tree::placeTrunk.
func (t *Tree) placeTrunk(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl) {
	// The base dirt block
	tx.AddBlockAt(x, y-1, z, block.VanillaDirt())
	for yy := 0; yy < trunkHeight; yy++ {
		if t.canOverride(tx.FetchBlockAt(x, y+yy, z)) {
			tx.AddBlockAt(x, y+yy, z, t.TrunkBlock.Clone())
		}
	}
}

// placeCanopy is a port of Tree::placeCanopy (the rounded blob of OakTree, BirchTree and
// JungleTree).
func (t *Tree) placeCanopy(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
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

func (t *Tree) canOverride(blk block.Behavior) bool {
	if t.canOverrideFn != nil {
		return t.canOverrideFn(blk)
	}
	return canOverride(blk)
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
	t := newTree(block.VanillaOakLog(), block.VanillaOakLeaves(), 7)
	t.beforeTransaction = func(random *utils.Random) { t.TreeHeight = random.NextBoundedInt(3) + 4 }
	return t
}

// NewBirchTree is a port of BirchTree. superBirch is its taller variant.
func NewBirchTree(superBirch bool) *Tree {
	t := newTree(block.VanillaBirchLog(), block.VanillaBirchLeaves(), 7)
	t.beforeTransaction = func(random *utils.Random) {
		t.TreeHeight = random.NextBoundedInt(3) + 5
		if superBirch {
			t.TreeHeight += 5
		}
	}
	return t
}

// NewJungleTree is a port of JungleTree.
func NewJungleTree() *Tree {
	return newTree(block.VanillaBlock("jungle_log"), block.VanillaBlock("jungle_leaves"), 8)
}

// NewSpruceTree is a port of SpruceTree.
func NewSpruceTree() *Tree {
	t := newTree(block.VanillaSpruceLog(), block.VanillaSpruceLeaves(), 10)
	t.beforeTransaction = func(random *utils.Random) { t.TreeHeight = random.NextBoundedInt(4) + 6 }
	t.generateTrunkHeight = func(random *utils.Random) int { return t.TreeHeight - random.NextBoundedInt(3) }
	t.placeCanopyFn = func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
		spruceCanopy(t, x, y, z, random, tx)
	}
	return t
}

// spruceCanopy is a port of SpruceTree::placeCanopy.
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

// horizontalFacings is Facing::HORIZONTAL.
var horizontalFacings = []math.Facing{math.North, math.South, math.West, math.East}

// acaciaMinHeight is AcaciaTree::MIN_HEIGHT.
const acaciaMinHeight = 5

// NewAcaciaTree is a port of AcaciaTree.
func NewAcaciaTree() *Tree {
	t := newTree(block.VanillaBlock("acacia_log"), block.VanillaBlock("acacia_leaves"), 0) // everything is overridden
	var mainBranchTip, secondBranchTip *math.Vector3

	placeBranch := func(tx *block.BlockTransactionImpl, start math.Vector3, branchFacing math.Facing, maxDiagonal, length int) math.Vector3 {
		diagonalPlaced := 0
		nextBlockPos := start
		for yy := 0; yy < length; yy++ {
			nextBlockPos = nextBlockPos.Up(1)
			if diagonalPlaced < maxDiagonal {
				nextBlockPos = nextBlockPos.GetSide(branchFacing, 1)
				diagonalPlaced++
			}
			tx.AddBlockAt(nextBlockPos.FloorX(), nextBlockPos.FloorY(), nextBlockPos.FloorZ(), t.TrunkBlock.Clone())
		}
		return nextBlockPos
	}
	placeCanopyLayer := func(tx *block.BlockTransactionImpl, center math.Vector3, radius, maxTaxicabDistance int) {
		centerX, centerY, centerZ := center.FloorX(), center.FloorY(), center.FloorZ()
		for x := centerX - radius; x <= centerX+radius; x++ {
			for z := centerZ - radius; z <= centerZ+radius; z++ {
				if absInt(x-centerX)+absInt(z-centerZ) <= maxTaxicabDistance && tx.FetchBlockAt(x, centerY, z).CanBeReplaced() {
					tx.AddBlockAt(x, centerY, z, t.LeafBlock.Clone())
				}
			}
		}
	}

	t.beforeTransaction = func(*utils.Random) { mainBranchTip, secondBranchTip = nil, nil }
	t.generateTrunkHeight = func(random *utils.Random) int {
		// 50% chance of 2 extra blocks, 33% chance 1 or 3, 17% chance 0 or 4
		return acaciaMinHeight + random.NextRange(0, 2) + random.NextRange(0, 2)
	}
	t.placeTrunkFn = func(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl) {
		// The base dirt block
		tx.AddBlockAt(x, y-1, z, block.VanillaDirt())
		firstBranchHeight := trunkHeight - 1 - random.NextRange(0, 3)
		for yy := 0; yy <= firstBranchHeight; yy++ {
			tx.AddBlockAt(x, y+yy, z, t.TrunkBlock.Clone())
		}
		// PHP picks the facings with array_rand (mt_rand), not the generation Random.
		mainBranchFacing := horizontalFacings[rand.Intn(len(horizontalFacings))]
		// this branch may grow a second trunk if the diagonal length is less than the max length
		tip := placeBranch(tx, math.NewVector3(float64(x), float64(y+firstBranchHeight), float64(z)), mainBranchFacing, random.NextRange(1, 3), trunkHeight-firstBranchHeight)
		mainBranchTip = &tip

		secondBranchFacing := horizontalFacings[rand.Intn(len(horizontalFacings))]
		if secondBranchFacing != mainBranchFacing {
			secondBranchLength := random.NextRange(1, 3)
			tip := placeBranch(tx, math.NewVector3(float64(x), float64(y+(firstBranchHeight-random.NextRange(0, 1))), float64(z)), secondBranchFacing, secondBranchLength, secondBranchLength) // the secondary branch may not form a second trunk
			secondBranchTip = &tip
		}
	}
	t.placeCanopyFn = func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
		if mainBranchTip != nil {
			placeCanopyLayer(tx, *mainBranchTip, 3, 5)
			placeCanopyLayer(tx, mainBranchTip.Up(1), 2, 2)
		}
		if secondBranchTip != nil {
			placeCanopyLayer(tx, *secondBranchTip, 2, 3)
			placeCanopyLayer(tx, secondBranchTip.Up(1), 1, 2)
		}
	}
	return t
}

// AzaleaTree constants, a port of AzaleaTree's private constants.
const (
	azaleaTreeHeightBase     = 4
	azaleaTreeHeightRandom   = 3
	azaleaMinHeightForLeaves = 3
	azaleaLeadupSmall        = 2
	azaleaLeadupLarge        = 3
	azaleaSideUpSteps        = 2
)

// NewAzaleaTree is a port of AzaleaTree.
func NewAzaleaTree() *Tree {
	t := newTree(block.VanillaOakLog(), block.VanillaBlock("azalea_leaves"), 0)
	var foliageAttachments []math.Vector3

	t.beforeTransaction = func(random *utils.Random) {
		t.TreeHeight = random.NextBoundedInt(azaleaTreeHeightRandom) + azaleaTreeHeightBase
		foliageAttachments = nil
	}
	t.generateTrunkHeight = func(random *utils.Random) int {
		return min(azaleaTreeHeightBase+random.NextBoundedInt(2), 5)
	}
	t.canOverrideFn = func(blk block.Behavior) bool {
		return canOverride(blk) || blk.GetTypeId() == block.AZALEA || blk.GetTypeId() == block.FLOWERING_AZALEA
	}
	t.placeTrunkFn = func(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl) {
		rootedDirt := block.VanillaDirt().(*block.Dirt)
		rootedDirt.SetDirtType(blockutils.DirtTypeRooted)
		tx.AddBlockAt(x, y-1, z, rootedDirt)

		direction := horizontalFacings[random.NextRange(0, len(horizontalFacings)-1)]
		offset := math.FacingOffset[direction]
		cx, cy, cz := x, y, z

		sideUpCount := min(azaleaSideUpSteps, max(0, trunkHeight-azaleaLeadupSmall))
		leadUpCount := min(trunkHeight-sideUpCount, azaleaLeadupLarge)
		total := leadUpCount + sideUpCount + 1
		if total < trunkHeight {
			leadUpCount += trunkHeight - total
		}
		for i := 0; i < total; i++ {
			isLeadUp := i < leadUpCount
			isSideUp := i < leadUpCount+sideUpCount-1
			if !isLeadUp {
				cx += offset[0]
				cz += offset[2]
			}
			if t.canOverride(tx.FetchBlockAt(cx, cy, cz)) {
				tx.AddBlockAt(cx, cy, cz, t.TrunkBlock.Clone())
			}
			if i >= azaleaMinHeightForLeaves {
				foliageAttachments = append(foliageAttachments, math.NewVector3(float64(cx), float64(cy), float64(cz)))
			}
			if isLeadUp || isSideUp {
				cy++
			}
		}
	}
	t.placeCanopyFn = func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
		radius, foliageHeight, attempts := 3, 2, 50
		visited := map[[3]int]bool{}
		for _, attachment := range foliageAttachments {
			centerX, centerY, centerZ := attachment.FloorX(), attachment.FloorY(), attachment.FloorZ()
			for a := 0; a < attempts; a++ {
				dx := random.NextBoundedInt(radius) - random.NextBoundedInt(radius)
				dy := random.NextBoundedInt(foliageHeight) - random.NextBoundedInt(foliageHeight)
				dz := random.NextBoundedInt(radius) - random.NextBoundedInt(radius)
				key := [3]int{centerX + dx, centerY + dy, centerZ + dz}
				if visited[key] {
					continue
				}
				visited[key] = true
				if tx.FetchBlockAt(key[0], key[1], key[2]).IsTransparent() {
					leafBlock := block.VanillaBlock("azalea_leaves")
					if random.NextBoundedInt(4) == 0 {
						leafBlock = block.VanillaBlock("flowering_azalea_leaves")
					}
					tx.AddBlockAt(key[0], key[1], key[2], leafBlock)
				}
			}
		}
	}
	return t
}

// NewNetherTree is a port of NetherTree (crimson and warped fungi): stemBlock, hatBlock (wart
// block), decorBlock (shroomlight).
func NewNetherTree(stemBlock, hatBlock, decorBlock block.Behavior, treeHeight int, hasVines, huge bool) *Tree {
	t := newTree(stemBlock, hatBlock, treeHeight)

	netherCanOverride := func(blk block.Behavior) bool {
		if blk.CanBeReplaced() {
			return true
		}
		tagged, ok := blk.(interface{ HasTypeTag(tag string) bool })
		return ok && tagged.HasTypeTag(block.BlockTypeTagsHugeFungusReplaceable)
	}
	t.canOverrideFn = netherCanOverride
	t.canPlaceObjectFn = func(block.World, int, int, int, *utils.Random) bool { return true }
	t.generateTrunkHeight = func(*utils.Random) int { return t.TreeHeight }

	tryPlaceWeepingVines := func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
		currentY := y - 1
		if !netherCanOverride(tx.FetchBlockAt(x, currentY, z)) {
			return
		}
		i := random.NextBoundedInt(5) + 1
		if random.NextBoundedInt(7) == 0 {
			i *= 2
		}
		maxAge := block.NetherVinesMaxAge
		startAge := 23
		for v := 0; v < i; v++ {
			vy := currentY - v
			if vy < 0 {
				break
			}
			if !netherCanOverride(tx.FetchBlockAt(x, vy, z)) {
				break
			}
			vines := block.VanillaBlock("weeping_vines").(*block.NetherVines)
			vines.SetAge(min(maxAge, startAge+v))
			tx.AddBlockAt(x, vy, z, vines)
		}
	}
	placeHatBlock := func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl, decorChance, hatChance, vineChance float64) {
		if random.NextFloat() < decorChance {
			tx.AddBlockAt(x, y, z, decorBlock.Clone())
		} else if random.NextFloat() < hatChance {
			tx.AddBlockAt(x, y, z, t.LeafBlock.Clone())
			if random.NextFloat() < vineChance {
				tryPlaceWeepingVines(x, y, z, random, tx)
			}
		}
	}
	placeHatDropBlock := func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl, isCrimson bool) {
		if tx.FetchBlockAt(x, y-1, z).GetTypeId() == t.LeafBlock.GetTypeId() {
			tx.AddBlockAt(x, y, z, t.LeafBlock.Clone())
		} else if random.NextFloat() < 0.15 {
			tx.AddBlockAt(x, y, z, t.LeafBlock.Clone())
			if isCrimson && random.NextBoundedInt(11) == 0 {
				tryPlaceWeepingVines(x, y, z, random, tx)
			}
		}
	}

	t.placeTrunkFn = func(x, y, z int, random *utils.Random, trunkHeight int, tx *block.BlockTransactionImpl) {
		i := 0
		if huge {
			i = 1
		}
		for j := -i; j <= i; j++ {
			for k := -i; k <= i; k++ {
				isCorner := huge && absInt(j) == i && absInt(k) == i
				for l := 0; l < trunkHeight; l++ {
					blockX, blockY, blockZ := x+j, y+l, z+k
					if (!isCorner || random.NextFloat() < 0.1) && netherCanOverride(tx.FetchBlockAt(blockX, blockY, blockZ)) {
						tx.AddBlockAt(blockX, blockY, blockZ, t.TrunkBlock.Clone())
					}
				}
			}
		}
	}
	t.placeCanopyFn = func(x, y, z int, random *utils.Random, tx *block.BlockTransactionImpl) {
		isCrimson := hasVines
		i := min(random.NextBoundedInt(1+t.TreeHeight/3)+5, t.TreeHeight)
		j := t.TreeHeight - i
		for k := j; k <= t.TreeHeight; k++ {
			l := 1
			if k < t.TreeHeight-random.NextBoundedInt(3) {
				l = 2
			}
			if i > 8 && k < j+4 {
				l = 3
			}
			if huge {
				l++
			}
			for i1 := -l; i1 <= l; i1++ {
				for j1 := -l; j1 <= l; j1++ {
					isEdgeX := i1 == -l || i1 == l
					isEdgeZ := j1 == -l || j1 == l
					isInner := !isEdgeX && !isEdgeZ && k != t.TreeHeight
					isCorner := isEdgeX && isEdgeZ
					isLowerSection := k < j+3

					blockX, blockY, blockZ := x+i1, y+k, z+j1
					if !netherCanOverride(tx.FetchBlockAt(blockX, blockY, blockZ)) {
						continue
					}
					vine := 0.0
					switch {
					case isLowerSection:
						if !isInner {
							placeHatDropBlock(blockX, blockY, blockZ, random, tx, isCrimson)
						}
					case isInner:
						if isCrimson {
							vine = 0.1
						}
						placeHatBlock(blockX, blockY, blockZ, random, tx, 0.1, 0.2, vine)
					case isCorner:
						if isCrimson {
							vine = 0.083
						}
						placeHatBlock(blockX, blockY, blockZ, random, tx, 0.01, 0.7, vine)
					default:
						if isCrimson {
							vine = 0.07
						}
						placeHatBlock(blockX, blockY, blockZ, random, tx, 0.0005, 0.98, vine)
					}
				}
			}
		}
	}
	return t
}
