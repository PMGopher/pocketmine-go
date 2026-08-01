package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CeilingCenterHangingSign is a port of pocketmine\block\CeilingCenterHangingSign.
type CeilingCenterHangingSign struct {
	BaseSign
	SignLikeRotationComponent
}

func NewCeilingCenterHangingSign(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *CeilingCenterHangingSign {
	c := &CeilingCenterHangingSign{BaseSign: BaseSign{
		Transparent:       Transparent{NewBlock(idInfo, name, typeInfo)},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}}
	c.Init(c)
	return c
}

func (c *CeilingCenterHangingSign) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CeilingCenterHangingSign) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeRotation(w)
}

func (c *CeilingCenterHangingSign) GetSupportingFace() math.Facing { return math.Up }

func (c *CeilingCenterHangingSign) GetFacingDegrees() float64 { return float64(c.Rotation) * 22.5 }

func (c *CeilingCenterHangingSign) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Down {
		return false
	}
	if player != nil {
		c.Rotation = SignLikeRotationFromYaw(player.GetYaw())
	}
	return c.BaseSign.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// canBeSupportedAt is a port of CeilingCenterHangingSign::canBeSupportedAt (a port of
// StaticSupportTrait's abstract method for this type - see Flower's doc comment for why this
// inlines the trait rather than sharing infrastructure for it).
func (c *CeilingCenterHangingSign) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Up, 1)
	return support.GetSupportType(math.Down).HasCenterSupport() || support.(blockGeometry).HasTypeTag(BlockTypeTagsHangingSign)
}

func (c *CeilingCenterHangingSign) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

// OnNearbyBlockChange overrides BaseSign's version - StaticSupportTrait's onNearbyBlockChange
// takes precedence over BaseSign's own in the PHP original (trait methods win over inherited
// parent methods).
func (c *CeilingCenterHangingSign) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	}
}
