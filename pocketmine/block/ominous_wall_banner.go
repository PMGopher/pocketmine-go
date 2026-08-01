package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// OminousWallBanner is a port of pocketmine\block\OminousWallBanner.
type OminousWallBanner struct {
	BaseOminousBanner
	HorizontalFacingComponent
}

func NewOminousWallBanner(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *OminousWallBanner {
	o := &OminousWallBanner{
		BaseOminousBanner:         BaseOminousBanner{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	o.Init(o)
	return o
}

func (o *OminousWallBanner) Clone() Behavior {
	c := *o
	c.rebind(&c)
	return &c
}

func (o *OminousWallBanner) DescribeBlockOnlyState(d runtime.DataDescriber) {
	o.DescribeHorizontalFacing(d)
}

func (o *OminousWallBanner) GetSupportingFace() math.Facing { return math.Opposite(o.Facing) }

func (o *OminousWallBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if math.FacingAxis(face) == math.AxisY {
		return false
	}
	o.Facing = face
	return o.BaseOminousBanner.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
