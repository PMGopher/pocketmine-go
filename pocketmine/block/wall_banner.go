package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// WallBanner is a port of pocketmine\block\WallBanner.
type WallBanner struct {
	BaseBanner
	HorizontalFacingComponent
}

func NewWallBanner(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WallBanner {
	w := &WallBanner{
		BaseBanner:                BaseBanner{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	w.Init(w)
	return w
}

func (w *WallBanner) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WallBanner) DescribeBlockOnlyState(d runtime.DataDescriber) { w.DescribeHorizontalFacing(d) }

func (w *WallBanner) DescribeBlockItemState(d runtime.DataDescriber) { w.DescribeColor(d) }

func (w *WallBanner) GetSupportingFace() math.Facing { return math.Opposite(w.Facing) }

func (w *WallBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if math.FacingAxis(face) == math.AxisY {
		return false
	}
	w.Facing = face
	return w.BaseBanner.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
