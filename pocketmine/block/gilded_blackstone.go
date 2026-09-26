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

// GetDropsForCompatibleTool is a port of GildedBlackstone::getDropsForCompatibleTool.
func (g *GildedBlackstone) GetDropsForCompatibleTool(item Item) []Item {
	if FortuneBonusChanceDivisor(item, 10, 3) {
		if nuggets := vanillaItemCount("gold_nugget", mtRand(2, 5)); nuggets != nil {
			return []Item{nuggets}
		}
	}
	return g.Block.GetDropsForCompatibleTool(item)
}

func (g *GildedBlackstone) IsAffectedBySilkTouch() bool { return true }
