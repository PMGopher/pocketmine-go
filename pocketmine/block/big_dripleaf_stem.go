package block

import "pocketmine-go/pocketmine/math"

// BigDripleafStem is a port of pocketmine\block\BigDripleafStem.
//
// AsItem should return VanillaBlocks.BIG_DRIPLEAF_HEAD().AsItem() — needs the unported block
// registry (see Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default
// for now.
type BigDripleafStem struct {
	BaseBigDripleaf
}

func NewBigDripleafStem(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BigDripleafStem {
	b := &BigDripleafStem{BaseBigDripleaf{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}}
	b.Init(b)
	return b
}

func (b *BigDripleafStem) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BigDripleafStem) IsHead() bool { return false }

func (b *BigDripleafStem) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }
