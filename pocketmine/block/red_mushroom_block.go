package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// RedMushroomBlock is a port of pocketmine\block\RedMushroomBlock.
type RedMushroomBlock struct {
	Opaque

	MushroomBlockTypeValue blockutils.MushroomBlockType
}

func NewRedMushroomBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedMushroomBlock {
	r := &RedMushroomBlock{
		Opaque:                 Opaque{NewBlock(idInfo, name, typeInfo)},
		MushroomBlockTypeValue: blockutils.MushroomBlockTypeAllCap,
	}
	r.Init(r)
	return r
}

func (r *RedMushroomBlock) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedMushroomBlock) DescribeBlockItemState(w runtime.DataDescriber) {
	t := int(r.MushroomBlockTypeValue)
	w.BoundedIntAuto(int(blockutils.MushroomBlockTypePores), int(blockutils.MushroomBlockTypeAllCap), &t)
	r.MushroomBlockTypeValue = blockutils.MushroomBlockType(t)
}

func (r *RedMushroomBlock) GetMushroomBlockType() blockutils.MushroomBlockType {
	return r.MushroomBlockTypeValue
}

func (r *RedMushroomBlock) SetMushroomBlockType(mushroomBlockType blockutils.MushroomBlockType) {
	r.MushroomBlockTypeValue = mushroomBlockType
}

func (r *RedMushroomBlock) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool should return [VanillaBlocks.RED_MUSHROOM().AsItem().SetCount(mt_rand(0,2))]
// — needs the unported block registry and real Item construction (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
//
// GetSilkTouchDrops/GetPickedItem have the same gap and already default to nil on Block, so
// there's nothing to override for them here.
func (r *RedMushroomBlock) GetDropsForCompatibleTool(item Item) []Item { return nil }
