package block

// IronOre is a port of pocketmine\block\IronOre.
type IronOre struct {
	Opaque
}

func NewIronOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *IronOre {
	i := &IronOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	i.Init(i)
	return i
}

func (i *IronOre) Clone() Behavior {
	c := *i
	c.rebind(&c)
	return &c
}

func (i *IronOre) IsAffectedBySilkTouch() bool { return true }
