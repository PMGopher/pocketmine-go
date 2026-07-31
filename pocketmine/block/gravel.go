package block

// Gravel is a port of pocketmine\block\Gravel.
type Gravel struct {
	Opaque
	FallableComponent
}

func NewGravel(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Gravel {
	g := &Gravel{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *Gravel) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

// OnNearbyBlockChange should replace itself with air and spawn a FallingBlock entity when
// unsupported — see Sand.OnNearbyBlockChange / FallableComponent's doc comment for the same gap.
func (g *Gravel) OnNearbyBlockChange() {}

// GetDropsForCompatibleTool's FortuneDropHelper-based flint chance needs the unported item
// package for real Item construction (see Block.GetDropsForCompatibleTool's doc comment); the
// fallback (no flint) path is fully portable, so that's what always runs for now.
func (g *Gravel) GetDropsForCompatibleTool(item Item) []Item {
	return g.Block.GetDropsForCompatibleTool(item)
}

func (g *Gravel) IsAffectedBySilkTouch() bool { return true }
