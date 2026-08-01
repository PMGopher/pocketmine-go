package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// BrownMushroomBlock is a port of pocketmine\block\BrownMushroomBlock.
type BrownMushroomBlock struct {
	RedMushroomBlock
}

func NewBrownMushroomBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BrownMushroomBlock {
	b := &BrownMushroomBlock{
		RedMushroomBlock: RedMushroomBlock{
			Opaque:                 Opaque{NewBlock(idInfo, name, typeInfo)},
			MushroomBlockTypeValue: blockutils.MushroomBlockTypeAllCap,
		},
	}
	b.Init(b)
	return b
}

func (b *BrownMushroomBlock) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool should return [VanillaBlocks.BROWN_MUSHROOM().AsItem().SetCount(mt_rand(0,2))]
// — same gap as RedMushroomBlock.GetDropsForCompatibleTool's doc comment.
func (b *BrownMushroomBlock) GetDropsForCompatibleTool(item Item) []Item { return nil }
