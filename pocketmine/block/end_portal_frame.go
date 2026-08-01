package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// EndPortalFrame is a port of pocketmine\block\EndPortalFrame.
type EndPortalFrame struct {
	Opaque
	HorizontalFacingComponent

	Eye bool
}

func NewEndPortalFrame(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *EndPortalFrame {
	e := &EndPortalFrame{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	e.Init(e)
	return e
}

func (e *EndPortalFrame) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *EndPortalFrame) DescribeBlockOnlyState(w runtime.DataDescriber) {
	e.DescribeHorizontalFacing(w)
	w.Bool(&e.Eye)
}

func (e *EndPortalFrame) HasEye() bool { return e.Eye }

func (e *EndPortalFrame) SetEye(eye bool) { e.Eye = eye }

func (e *EndPortalFrame) GetLightLevel() int { return 1 }

func (e *EndPortalFrame) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 3.0/16)}
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (e *EndPortalFrame) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		e.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return e.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
