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

// GetDropsForCompatibleTool should return glowstone dust scaled via FortuneDropHelper — needs
// real Item construction from the unported item package (see Block.GetDropsForCompatibleTool's
// doc comment), so this returns nil for now.
func (g *Glowstone) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (g *Glowstone) IsAffectedBySilkTouch() bool { return true }
