package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperGrate is a port of pocketmine\block\CopperGrate.
//
// Waterlogging is a TODO in the PHP original too.
type CopperGrate struct {
	Transparent
	CopperComponent
}

func NewCopperGrate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperGrate {
	c := &CopperGrate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CopperGrate) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperGrate) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperGrate) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
