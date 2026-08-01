package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CeilingEdgesHangingSign is a port of pocketmine\block\CeilingEdgesHangingSign.
type CeilingEdgesHangingSign struct {
	BaseSign
	HorizontalFacingComponent
}

func NewCeilingEdgesHangingSign(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *CeilingEdgesHangingSign {
	c := &CeilingEdgesHangingSign{
		BaseSign: BaseSign{
			Transparent:       Transparent{NewBlock(idInfo, name, typeInfo)},
			WoodTypeComponent: NewWoodTypeComponent(woodType),
		},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *CeilingEdgesHangingSign) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CeilingEdgesHangingSign) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
}

func (c *CeilingEdgesHangingSign) GetSupportingFace() math.Facing { return math.Up }

func (c *CeilingEdgesHangingSign) GetFacingDegrees() float64 {
	switch c.Facing {
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

// canBeSupportedAt is a port of CeilingEdgesHangingSign::canBeSupportedAt.
func (c *CeilingEdgesHangingSign) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Up, 1)
	if support.GetSupportType(math.Down) == blockutils.SupportTypeFull {
		return true
	}
	var otherFacing math.Facing
	switch s := support.(type) {
	case *WallHangingSign:
		otherFacing = s.Facing
	case *CeilingEdgesHangingSign:
		otherFacing = s.Facing
	default:
		return false
	}
	return math.FacingAxis(otherFacing) == math.FacingAxis(c.Facing)
}

func (c *CeilingEdgesHangingSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Down {
		return false
	}
	if player != nil {
		c.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	if !c.canBeSupportedAt(blockReplace) {
		return false
	}
	return c.BaseSign.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (c *CeilingEdgesHangingSign) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	}
}
