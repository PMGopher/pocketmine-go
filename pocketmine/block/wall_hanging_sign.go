package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// WallHangingSign is a port of pocketmine\block\WallHangingSign.
type WallHangingSign struct {
	BaseSign
	HorizontalFacingComponent
}

func NewWallHangingSign(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WallHangingSign {
	w := &WallHangingSign{
		BaseSign: BaseSign{
			Transparent:       Transparent{NewBlock(idInfo, name, typeInfo)},
			WoodTypeComponent: NewWoodTypeComponent(woodType),
		},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	w.Init(w)
	return w
}

func (w *WallHangingSign) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WallHangingSign) DescribeBlockOnlyState(d runtime.DataDescriber) {
	w.DescribeHorizontalFacing(d)
}

func (w *WallHangingSign) GetSupportingFace() math.Facing { return math.RotateY(w.Facing, true) }

// OnNearbyBlockChange is a port of WallHangingSign::onNearbyBlockChange - a deliberate no-op,
// disabling BaseSign's default self-destruct behaviour (matching the PHP original's own comment).
func (w *WallHangingSign) OnNearbyBlockChange() {}

func (w *WallHangingSign) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	// Only the cross bar is collidable.
	bb := math.OneAABB().TrimmedCopy(math.Down, 7.0/8).SquashedCopy(math.FacingAxis(w.Facing), 3.0/4)
	return []math.AxisAlignedBB{bb}
}

// canBeSupportedAt is a port of WallHangingSign::canBeSupportedAt.
func (w *WallHangingSign) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	if other, ok := blk.(*WallHangingSign); ok {
		if math.FacingAxis(math.RotateY(other.Facing, true)) == math.FacingAxis(face) {
			return true
		}
	}
	return blk.GetSupportType(math.Opposite(face)) == blockutils.SupportTypeFull
}

func (w *WallHangingSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player == nil {
		return false
	}
	attachFace := face
	if math.FacingAxis(face) == math.AxisY {
		attachFace = math.RotateY(player.GetHorizontalFacing(), true)
	}

	geo := blockReplace.(blockGeometry)
	var direction math.Facing
	if w.canBeSupportedAt(geo.GetSide(attachFace, 1), attachFace) {
		direction = attachFace
	} else if opposite := math.Opposite(attachFace); w.canBeSupportedAt(geo.GetSide(opposite, 1), opposite) {
		direction = opposite
	} else {
		return false
	}

	w.Facing = math.RotateY(math.Opposite(direction), true)
	// The front should always face the player if possible.
	if w.Facing == player.GetHorizontalFacing() {
		w.Facing = math.Opposite(w.Facing)
	}

	return w.BaseSign.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (w *WallHangingSign) GetFacingDegrees() float64 {
	switch w.Facing {
	case math.South:
		return 0
	case math.West:
		return 90
	case math.North:
		return 180
	case math.East:
		return 270
	default:
		panic("Invalid facing direction")
	}
}
