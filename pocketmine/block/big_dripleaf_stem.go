package block

import "pocketmine-go/pocketmine/math"

// BigDripleafStem is a port of pocketmine\block\BigDripleafStem.
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

// GetDropsForCompatibleTool/GetPickedItem port BigDripleafStem::asItem()'s effect: this block
// drops/is picked as a Big Dripleaf (the head), matching
// `VanillaBlocks::BIG_DRIPLEAF_HEAD()->asItem()`. AsItem() itself isn't self-dispatched in this
// port (see Block.AsItem's doc comment), so both call sites are overridden directly here instead
// of overriding AsItem().
func (b *BigDripleafStem) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaBigDripleafHead())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}

func (b *BigDripleafStem) GetPickedItem(addUserData bool) Item {
	return asItemOrNil(VanillaBigDripleafHead())
}
