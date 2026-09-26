package block

// NetherReactor is a port of pocketmine\block\NetherReactor.
type NetherReactor struct {
	Opaque
}

func NewNetherReactor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherReactor {
	n := &NetherReactor{Opaque{NewBlock(idInfo, name, typeInfo)}}
	n.Init(n)
	return n
}

func (n *NetherReactor) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool is a port of NetherReactor::getDropsForCompatibleTool.
func (n *NetherReactor) GetDropsForCompatibleTool(item Item) []Item {
	var drops []Item
	if iron := vanillaItemCount("iron_ingot", 6); iron != nil {
		drops = append(drops, iron)
	}
	if diamonds := vanillaItemCount("diamond", 3); diamonds != nil {
		drops = append(drops, diamonds)
	}
	return drops
}
