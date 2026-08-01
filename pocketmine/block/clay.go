package block

// Clay is a port of pocketmine\block\Clay.
type Clay struct {
	Opaque
}

func NewClay(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Clay {
	c := &Clay{Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *Clay) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

// GetDropsForCompatibleTool should return 4 clay balls — needs real Item construction from the
// unported item package (see Block.GetDropsForCompatibleTool's doc comment), so this returns nil
// for now.
func (c *Clay) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (c *Clay) IsAffectedBySilkTouch() bool { return true }
