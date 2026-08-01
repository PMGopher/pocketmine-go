package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperSlab is a port of pocketmine\block\CopperSlab.
type CopperSlab struct {
	Slab
	CopperComponent
}

func NewCopperSlab(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperSlab {
	c := &CopperSlab{Slab: Slab{Transparent: Transparent{NewBlock(idInfo, name+" Slab", typeInfo)}, SlabTypeValue: blockutils.SlabTypeBottom}}
	c.Init(c)
	return c
}

func (c *CopperSlab) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperSlab) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperSlab) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
