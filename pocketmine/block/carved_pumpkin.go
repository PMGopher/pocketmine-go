package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CarvedPumpkin is a port of pocketmine\block\CarvedPumpkin.
type CarvedPumpkin struct {
	Opaque
	HorizontalFacingComponent
}

func NewCarvedPumpkin(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CarvedPumpkin {
	c := &CarvedPumpkin{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *CarvedPumpkin) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CarvedPumpkin) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (c *CarvedPumpkin) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		c.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
