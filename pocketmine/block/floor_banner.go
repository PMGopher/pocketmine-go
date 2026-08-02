package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// FloorBanner is a port of pocketmine\block\FloorBanner.
type FloorBanner struct {
	BaseBanner
	SignLikeRotationComponent
}

func NewFloorBanner(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *FloorBanner {
	f := &FloorBanner{BaseBanner: BaseBanner{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}}
	f.Init(f)
	return f
}

func (f *FloorBanner) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FloorBanner) DescribeBlockOnlyState(w runtime.DataDescriber) { f.DescribeRotation(w) }

func (f *FloorBanner) DescribeBlockItemState(w runtime.DataDescriber) { f.DescribeColor(w) }

func (f *FloorBanner) GetSupportingFace() math.Facing { return math.Down }

// GetOminousVersion is a port of FloorBanner::getOminousVersion.
func (f *FloorBanner) GetOminousVersion() Behavior {
	ominous := VanillaOminousBanner().(*OminousFloorBanner)
	ominous.SetRotation(f.Rotation)
	return ominous
}

func (f *FloorBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Up {
		return false
	}
	if player != nil {
		f.Rotation = SignLikeRotationFromYaw(player.GetYaw())
	}
	return f.BaseBanner.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
