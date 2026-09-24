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

// OnNearbyBlockChange is FallableTrait::onNearbyBlockChange.
func (g *Gravel) OnNearbyBlockChange() { FallableOnNearbyBlockChange(g.self) }

// GetDropsForCompatibleTool's FortuneDropHelper-based flint chance needs the unported item
// package for real Item construction (see Block.GetDropsForCompatibleTool's doc comment); the
// fallback (no flint) path is fully portable, so that's what always runs for now.
func (g *Gravel) GetDropsForCompatibleTool(item Item) []Item {
	return g.Block.GetDropsForCompatibleTool(item)
}

func (g *Gravel) IsAffectedBySilkTouch() bool { return true }
