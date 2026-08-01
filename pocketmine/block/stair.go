package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Stair is a port of pocketmine\block\Stair.
type Stair struct {
	Transparent
	HorizontalFacingComponent

	UpsideDown bool
	Shape      blockutils.StairShape
}

func NewStair(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Stair {
	s := &Stair{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		Shape:                     blockutils.StairShapeStraight,
	}
	s.Init(s)
	return s
}

func (s *Stair) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Stair) DescribeBlockOnlyState(w runtime.DataDescriber) {
	s.DescribeHorizontalFacing(w)
	w.Bool(&s.UpsideDown)
}

func (s *Stair) ReadStateFromWorld() Behavior {
	s.Block.ReadStateFromWorld()

	s.collisionBoxes = nil
	s.haveCollisionBoxes = false

	clockwise := math.RotateY(s.Facing, true)
	if backFacing, ok := s.getPossibleCornerFacing(false); ok {
		if backFacing == clockwise {
			s.Shape = blockutils.StairShapeOuterRight
		} else {
			s.Shape = blockutils.StairShapeOuterLeft
		}
	} else if frontFacing, ok := s.getPossibleCornerFacing(true); ok {
		if frontFacing == clockwise {
			s.Shape = blockutils.StairShapeInnerRight
		} else {
			s.Shape = blockutils.StairShapeInnerLeft
		}
	} else {
		s.Shape = blockutils.StairShapeStraight
	}

	return s.self
}

func (s *Stair) IsUpsideDown() bool { return s.UpsideDown }

func (s *Stair) SetUpsideDown(upsideDown bool) { s.UpsideDown = upsideDown }

func (s *Stair) GetShape() blockutils.StairShape { return s.Shape }

func (s *Stair) SetShape(shape blockutils.StairShape) { s.Shape = shape }

func (s *Stair) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	topStepFace := math.Up
	if s.UpsideDown {
		topStepFace = math.Down
	}

	bottom := math.OneAABB()
	bottom.Trim(topStepFace, 0.5)
	bbs := []math.AxisAlignedBB{bottom}

	topStep := math.OneAABB()
	topStep.Trim(math.Opposite(topStepFace), 0.5)
	topStep.Trim(math.Opposite(s.Facing), 0.5)

	switch s.Shape {
	case blockutils.StairShapeOuterLeft, blockutils.StairShapeOuterRight:
		topStep.Trim(math.RotateY(s.Facing, s.Shape == blockutils.StairShapeOuterLeft), 0.5)
	case blockutils.StairShapeInnerLeft, blockutils.StairShapeInnerRight:
		extra := math.OneAABB()
		extra.Trim(math.Opposite(topStepFace), 0.5)
		extra.Trim(s.Facing, 0.5) // avoid overlapping with main step
		extra.Trim(math.RotateY(s.Facing, s.Shape == blockutils.StairShapeInnerLeft), 0.5)
		bbs = append(bbs, extra)
	}

	bbs = append(bbs, topStep)

	return bbs
}

func (s *Stair) GetSupportType(facing math.Facing) blockutils.SupportType {
	if (facing == math.Up && s.UpsideDown) ||
		(facing == math.Down && !s.UpsideDown) ||
		(facing == s.Facing && s.Shape != blockutils.StairShapeOuterLeft && s.Shape != blockutils.StairShapeOuterRight) ||
		(facing == math.Rotate(s.Facing, math.AxisY, false) && s.Shape == blockutils.StairShapeInnerLeft) ||
		(facing == math.Rotate(s.Facing, math.AxisY, true) && s.Shape == blockutils.StairShapeInnerRight) {
		return blockutils.SupportTypeFull
	}
	return blockutils.SupportTypeNone
}

func (s *Stair) getPossibleCornerFacing(oppositeFacing bool) (math.Facing, bool) {
	checkFacing := s.Facing
	if oppositeFacing {
		checkFacing = math.Opposite(s.Facing)
	}
	side := s.GetSide(checkFacing, 1)
	other, ok := side.(*Stair)
	if ok && other.UpsideDown == s.UpsideDown && math.FacingAxis(other.Facing) != math.FacingAxis(s.Facing) {
		return other.Facing, true
	}
	return 0, false
}

func (s *Stair) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		s.Facing = player.GetHorizontalFacing()
	}
	s.UpsideDown = (clickVector.Y > 0.5 && face != math.Up) || face == math.Down
	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
