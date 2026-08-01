package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperStairs is a port of pocketmine\block\CopperStairs.
type CopperStairs struct {
	Stair
	CopperComponent
}

func NewCopperStairs(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperStairs {
	c := &CopperStairs{
		Stair: Stair{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
			Shape:                     blockutils.StairShapeStraight,
		},
	}
	c.Init(c)
	return c
}

func (c *CopperStairs) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperStairs) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperStairs) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
