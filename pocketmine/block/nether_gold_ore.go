package block

// NetherGoldOre is a port of pocketmine\block\NetherGoldOre.
type NetherGoldOre struct {
	Opaque
}

func NewNetherGoldOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherGoldOre {
	n := &NetherGoldOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	n.Init(n)
	return n
}

func (n *NetherGoldOre) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherGoldOre) IsAffectedBySilkTouch() bool { return true }
