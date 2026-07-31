package block

// Glass is a port of pocketmine\block\Glass.
type Glass struct {
	Transparent
}

func NewGlass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Glass {
	g := &Glass{Transparent{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *Glass) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool deliberately returns nothing — glass shatters instead of dropping
// itself, matching the PHP original's `return [];` (this isn't a not-yet-ported gap).
func (g *Glass) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (g *Glass) IsAffectedBySilkTouch() bool { return true }
