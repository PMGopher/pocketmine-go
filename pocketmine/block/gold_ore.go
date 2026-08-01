package block

// GoldOre is a port of pocketmine\block\GoldOre.
type GoldOre struct {
	Opaque
}

func NewGoldOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GoldOre {
	g := &GoldOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *GoldOre) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GoldOre) IsAffectedBySilkTouch() bool { return true }
