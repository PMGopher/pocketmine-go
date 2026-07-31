package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperChain is a port of pocketmine\block\CopperChain.
type CopperChain struct {
	Chain
	CopperComponent
}

func NewCopperChain(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperChain {
	c := &CopperChain{
		Chain: Chain{
			Transparent:             Transparent{NewBlock(idInfo, name, typeInfo)},
			PillarRotationComponent: NewPillarRotationComponent(),
		},
	}
	c.Init(c)
	return c
}

func (c *CopperChain) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperChain) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperChain) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
