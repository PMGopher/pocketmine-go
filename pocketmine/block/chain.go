package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Chain is a port of pocketmine\block\Chain.
type Chain struct {
	Transparent
	PillarRotationComponent
}

func NewChain(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Chain {
	c := &Chain{
		Transparent:             Transparent{NewBlock(idInfo, name, typeInfo)},
		PillarRotationComponent: NewPillarRotationComponent(),
	}
	c.Init(c)
	return c
}

func (c *Chain) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Chain) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeAxis(w) }

func (c *Chain) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	c.SetAxisFromFace(face)
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (c *Chain) GetSupportType(facing math.Facing) blockutils.SupportType {
	if c.Axis == math.AxisY && math.FacingAxis(facing) == math.AxisY {
		return blockutils.SupportTypeCenter
	}
	return blockutils.SupportTypeNone
}

func (c *Chain) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	bb := math.OneAABB()
	for _, axis := range []math.Axis{math.AxisY, math.AxisZ, math.AxisX} {
		if axis != c.Axis {
			bb = bb.SquashedCopy(axis, 13.0/32)
		}
	}
	return []math.AxisAlignedBB{bb}
}
