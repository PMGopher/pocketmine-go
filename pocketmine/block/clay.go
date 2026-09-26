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

// GetDropsForCompatibleTool is a port of Clay::getDropsForCompatibleTool.
func (c *Clay) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("clay", 4))
}

func (c *Clay) IsAffectedBySilkTouch() bool { return true }
