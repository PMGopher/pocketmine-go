package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// Concrete is a port of pocketmine\block\Concrete.
type Concrete struct {
	Opaque
	ColorComponent
}

func NewConcrete(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Concrete {
	c := &Concrete{
		Opaque:         Opaque{NewBlock(idInfo, name, typeInfo)},
		ColorComponent: NewColorComponent(),
	}
	c.Init(c)
	return c
}

func (c *Concrete) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Concrete) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeColor(w) }
