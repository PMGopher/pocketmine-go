package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// OminousFloorBanner is a port of pocketmine\block\OminousFloorBanner.
type OminousFloorBanner struct {
	BaseOminousBanner
	SignLikeRotationComponent
}

func NewOminousFloorBanner(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *OminousFloorBanner {
	o := &OminousFloorBanner{BaseOminousBanner: BaseOminousBanner{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}}
	o.Init(o)
	return o
}

func (o *OminousFloorBanner) Clone() Behavior {
	c := *o
	c.rebind(&c)
	return &c
}

func (o *OminousFloorBanner) DescribeBlockOnlyState(w runtime.DataDescriber) { o.DescribeRotation(w) }

func (o *OminousFloorBanner) GetSupportingFace() math.Facing { return math.Down }

func (o *OminousFloorBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Up {
		return false
	}
	if player != nil {
		o.Rotation = SignLikeRotationFromYaw(player.GetYaw())
	}
	return o.BaseOminousBanner.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
