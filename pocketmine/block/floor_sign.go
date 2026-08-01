package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// FloorSign is a port of pocketmine\block\FloorSign.
type FloorSign struct {
	BaseSign
	SignLikeRotationComponent
}

func NewFloorSign(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *FloorSign {
	f := &FloorSign{BaseSign: BaseSign{
		Transparent:       Transparent{NewBlock(idInfo, name, typeInfo)},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}}
	f.Init(f)
	return f
}

func (f *FloorSign) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FloorSign) DescribeBlockOnlyState(w runtime.DataDescriber) { f.DescribeRotation(w) }

func (f *FloorSign) GetSupportingFace() math.Facing { return math.Down }

func (f *FloorSign) GetFacingDegrees() float64 { return float64(f.Rotation) * 22.5 }

func (f *FloorSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Up {
		return false
	}
	if player != nil {
		f.Rotation = SignLikeRotationFromYaw(player.GetYaw())
	}
	return f.BaseSign.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
