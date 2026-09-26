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

// GetDropsForCompatibleTool is a port of RedMushroomBlock::getDropsForCompatibleTool.
func (r *RedMushroomBlock) GetDropsForCompatibleTool(item Item) []Item {
	return mushroomDrops("red_mushroom")
}

// mushroomDrops is VanillaBlocks::X_MUSHROOM()->asItem()->setCount(mt_rand(0, 2)).
func mushroomDrops(name string) []Item {
	drop := asItemOrNil(VanillaBlock(name))
	if drop == nil {
		return nil
	}
	drop.SetCount(mtRand(0, 2))
	return []Item{drop}
}

// GetPickedItem is a port of RedMushroomBlock::getPickedItem: always the all-cap variant.
func (r *RedMushroomBlock) GetPickedItem(addUserData bool) Item {
	allCap := r.self.Clone()
	allCap.(interface {
		SetMushroomBlockType(blockutils.MushroomBlockType)
	}).SetMushroomBlockType(blockutils.MushroomBlockTypeAllCap)
	return asItemOrNil(allCap)
}
