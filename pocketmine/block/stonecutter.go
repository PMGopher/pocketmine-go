package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Stonecutter is a port of pocketmine\block\Stonecutter.
type Stonecutter struct {
	Transparent
	HorizontalFacingComponent
}

func NewStonecutter(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Stonecutter {
	s := &Stonecutter{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	s.Init(s)
	return s
}

func (s *Stonecutter) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Stonecutter) DescribeBlockOnlyState(w runtime.DataDescriber) { s.DescribeHorizontalFacing(w) }

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (s *Stonecutter) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		s.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of Stonecutter::onInteract.
func (s *Stonecutter) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	openWindow(player, WindowStonecutter, s.position)
	return true
}

func (s *Stonecutter) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 7.0/16)}
}

func (s *Stonecutter) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
