package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Ladder is a port of pocketmine\block\Ladder.
type Ladder struct {
	Transparent
	HorizontalFacingComponent
}

func NewLadder(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Ladder {
	l := &Ladder{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	l.Init(l)
	return l
}

func (l *Ladder) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Ladder) DescribeBlockOnlyState(w runtime.DataDescriber) { l.DescribeHorizontalFacing(w) }

func (l *Ladder) HasEntityCollision() bool { return true }

func (l *Ladder) IsSolid() bool { return false }

func (l *Ladder) CanClimb() bool { return true }

func (l *Ladder) OnEntityInside(entity Entity) bool {
	if living, ok := entity.(Living); ok {
		if living.GetPosition().Floor().DistanceSquared(l.position.AsVector3()) < 1 {
			living.ResetFallDistance()
			living.SetOnGround(true)
		}
	}
	return true
}

func (l *Ladder) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(l.Facing, 13.0/16.0)}
}

func (l *Ladder) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (l *Ladder) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if l.canBeSupportedAt(blockReplace, math.Opposite(face)) && math.FacingAxis(face) != math.AxisY {
		l.Facing = face
		return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

func (l *Ladder) OnNearbyBlockChange() {
	if !l.canBeSupportedAt(l.self, math.Opposite(l.Facing)) {
		if world, err := l.position.GetWorld(); err == nil {
			world.UseBreakOn(l.position.AsVector3())
		}
	}
}

func (l *Ladder) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(face) == blockutils.SupportTypeFull
}
