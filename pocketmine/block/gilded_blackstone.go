package block

// GildedBlackstone is a port of pocketmine\block\GildedBlackstone.
type GildedBlackstone struct {
	Opaque
}

func NewGildedBlackstone(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GildedBlackstone {
	g := &GildedBlackstone{Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *GildedBlackstone) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool's gold-nugget bonus chance needs the unported item package for real
// Item construction (see Block.GetDropsForCompatibleTool's doc comment); the fallback path
// (parent::getDropsForCompatibleTool) is fully portable and always runs for now.
func (g *GildedBlackstone) GetDropsForCompatibleTool(item Item) []Item {
	return g.Block.GetDropsForCompatibleTool(item)
}

func (g *GildedBlackstone) IsAffectedBySilkTouch() bool { return true }
