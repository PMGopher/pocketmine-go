package object

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
)

// NewTreeFromType is a port of TreeFactory::get: the tree for a tree type, or nil if there is none
// (dark oak).
func NewTreeFromType(random *utils.Random, t TreeType) *Tree {
	switch t {
	case TreeTypeOak:
		return NewOakTree() //TODO: big oak has a 1/10 chance
	case TreeTypeSpruce:
		return NewSpruceTree()
	case TreeTypeJungle:
		return NewJungleTree()
	case TreeTypeAcacia:
		return NewAcaciaTree()
	case TreeTypeBirch:
		return NewBirchTree(random.NextBoundedInt(39) == 0)
	case TreeTypeAzalea:
		return NewAzaleaTree()
	case TreeTypeCrimson:
		return NewNetherTree(block.VanillaBlock("crimson_stem"), block.VanillaBlock("nether_wart_block"), block.VanillaBlock("shroomlight"), netherTreeHeight(random), true, random.NextFloat() < 0.06)
	case TreeTypeWarped:
		return NewNetherTree(block.VanillaBlock("warped_stem"), block.VanillaBlock("warped_wart_block"), block.VanillaBlock("shroomlight"), netherTreeHeight(random), false, random.NextFloat() < 0.06)
	}
	return nil
}

// netherTreeHeight is TreeFactory's ($random->nextBoundedInt(9) + 4) * ($random->nextBoundedInt(12) === 0 ? 2 : 1).
func netherTreeHeight(random *utils.Random) int {
	h := random.NextBoundedInt(9) + 4
	if random.NextBoundedInt(12) == 0 {
		h *= 2
	}
	return h
}

// init gives the block package tree growth (Sapling, Azalea and NetherFungus::grow use
// TreeFactory, which it can't import) and TallGrass::growGrass (Grass::onInteract).
func init() {
	block.TreeTransactionFunc = func(treeType block.TreeType, world block.World, x, y, z int, random *utils.Random) *block.BlockTransactionImpl {
		tree := NewTreeFromType(random, TreeType(treeType))
		if tree == nil {
			return nil
		}
		return tree.GetBlockTransaction(world, x, y, z, random)
	}
	block.GrowGrassFunc = GrowGrass
}
