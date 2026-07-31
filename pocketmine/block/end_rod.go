package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// EndRod is a port of pocketmine\block\EndRod.
type EndRod struct {
	Flowable
	FacingComponent
}

func NewEndRod(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *EndRod {
	e := &EndRod{
		Flowable:        Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		FacingComponent: NewFacingComponent(),
	}
	e.Init(e)
	return e
}

func (e *EndRod) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *EndRod) DescribeBlockOnlyState(w runtime.DataDescriber) {
	e.DescribeFacing(w)
}

func (e *EndRod) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	e.Facing = face
	if clicked, ok := blockClicked.(*EndRod); ok && clicked.Facing == e.Facing {
		e.Facing = math.Opposite(face)
	}
	return e.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (e *EndRod) IsSolid() bool { return true }

func (e *EndRod) GetLightLevel() int { return 14 }

func (e *EndRod) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	myAxis := math.FacingAxis(e.Facing)

	bb := math.OneAABB()
	for _, axis := range []math.Axis{math.AxisY, math.AxisZ, math.AxisX} {
		if axis == myAxis {
			continue
		}
		bb.Squash(axis, 6.0/16.0)
	}
	return []math.AxisAlignedBB{bb}
}
