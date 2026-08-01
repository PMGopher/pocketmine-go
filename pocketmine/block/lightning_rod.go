package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// LightningRod is a port of pocketmine\block\LightningRod.
type LightningRod struct {
	Transparent
	CopperComponent
	FacingComponent
}

func NewLightningRod(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *LightningRod {
	l := &LightningRod{
		Transparent:     Transparent{NewBlock(idInfo, name, typeInfo)},
		FacingComponent: NewFacingComponent(),
	}
	l.Init(l)
	return l
}

func (l *LightningRod) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *LightningRod) DescribeBlockOnlyState(w runtime.DataDescriber) { l.DescribeFacing(w) }

func (l *LightningRod) DescribeBlockItemState(w runtime.DataDescriber) { l.DescribeCopper(w) }

func (l *LightningRod) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	myAxis := math.FacingAxis(l.Facing)

	bb := math.OneAABB()
	for _, axis := range []math.Axis{math.AxisX, math.AxisY, math.AxisZ} {
		if axis != myAxis {
			bb.Squash(axis, 6.0/16.0)
		}
	}
	return []math.AxisAlignedBB{bb}
}

func (l *LightningRod) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	l.Facing = face
	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (l *LightningRod) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return l.OnInteractCopper(l.self, l.position, item)
}
