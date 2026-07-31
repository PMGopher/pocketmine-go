package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Copper is a port of pocketmine\block\Copper.
type Copper struct {
	Opaque
	CopperComponent
}

func NewCopper(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Copper {
	c := &Copper{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *Copper) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Copper) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *Copper) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
