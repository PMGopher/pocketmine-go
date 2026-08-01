package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperBars is a port of pocketmine\block\CopperBars.
type CopperBars struct {
	Thin
	CopperComponent
}

func NewCopperBars(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperBars {
	c := &CopperBars{Thin: Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}}
	c.Init(c)
	return c
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (c *CopperBars) Clone() Behavior {
	cl := *c
	cl.Connections = make(map[math.Facing]bool, len(c.Connections))
	for k, v := range c.Connections {
		cl.Connections[k] = v
	}
	cl.rebind(&cl)
	return &cl
}

func (c *CopperBars) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperBars) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
