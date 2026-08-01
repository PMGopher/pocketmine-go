package block

// Coal is a port of pocketmine\block\Coal.
type Coal struct {
	Opaque
}

func NewCoal(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Coal {
	c := &Coal{Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *Coal) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Coal) GetFuelTime() int { return 16000 }

func (c *Coal) GetFlameEncouragement() int { return 5 }

func (c *Coal) GetFlammability() int { return 5 }
