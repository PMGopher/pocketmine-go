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

// GetDropsForCompatibleTool is a port of Gravel::getDropsForCompatibleTool.
func (g *Gravel) GetDropsForCompatibleTool(item Item) []Item {
	if FortuneBonusChanceDivisor(item, 10, 3) {
		if flint := vanillaItem("flint"); flint != nil {
			return []Item{flint}
		}
	}
	return g.Block.GetDropsForCompatibleTool(item)
}

func (g *Gravel) IsAffectedBySilkTouch() bool { return true }
