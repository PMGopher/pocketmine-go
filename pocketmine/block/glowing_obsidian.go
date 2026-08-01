package block

// GlowingObsidian is a port of pocketmine\block\GlowingObsidian.
type GlowingObsidian struct {
	Opaque
}

func NewGlowingObsidian(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GlowingObsidian {
	g := &GlowingObsidian{Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *GlowingObsidian) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GlowingObsidian) GetLightLevel() int { return 12 }
