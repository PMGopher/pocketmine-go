package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// WallSign is a port of pocketmine\block\WallSign.
type WallSign struct {
	BaseSign
	HorizontalFacingComponent
}

func NewWallSign(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WallSign {
	w := &WallSign{
		BaseSign: BaseSign{
			Transparent:       Transparent{NewBlock(idInfo, name, typeInfo)},
			WoodTypeComponent: NewWoodTypeComponent(woodType),
		},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	w.Init(w)
	return w
}

func (w *WallSign) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WallSign) DescribeBlockOnlyState(d runtime.DataDescriber) { w.DescribeHorizontalFacing(d) }

func (w *WallSign) GetSupportingFace() math.Facing { return math.Opposite(w.Facing) }

// getHitboxCenter panics for a non-horizontal facing, mirroring the PHP original's
// AssumptionFailedError (a programmer error / corrupted state, not a normal runtime condition).
func (w *WallSign) getHitboxCenter() math.Vector3 {
	var xOffset, zOffset float64
	switch w.Facing {
	case math.North:
		xOffset, zOffset = 0, 15.0/16
	case math.South:
		xOffset, zOffset = 0, 1.0/16
	case math.West:
		xOffset, zOffset = 15.0/16, 0
	case math.East:
		xOffset, zOffset = 1.0/16, 0
	default:
		panic("Invalid facing direction")
	}
	return w.position.AsVector3().Add(xOffset, 0.5, zOffset)
}

func (w *WallSign) GetFacingDegrees() float64 {
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

func (w *WallSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if math.FacingAxis(face) == math.AxisY {
		return false
	}
	w.Facing = face
	return w.BaseSign.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
