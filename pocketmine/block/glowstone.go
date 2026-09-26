package block

// Glowstone is a port of pocketmine\block\Glowstone.
type Glowstone struct {
	Transparent
}

func NewGlowstone(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Glowstone {
	g := &Glowstone{Transparent{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *Glowstone) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *Glowstone) GetLightLevel() int { return 15 }

// GetDropsForCompatibleTool is a port of Glowstone::getDropsForCompatibleTool.
func (g *Glowstone) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("glowstone_dust", min(4, FortuneDiscrete(item, 2, 4))))
}

func (g *Glowstone) IsAffectedBySilkTouch() bool { return true }
