package block

// CopperOre is a port of pocketmine\block\CopperOre.
type CopperOre struct {
	Opaque
}

func NewCopperOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperOre {
	c := &CopperOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CopperOre) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperOre) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool's FortuneDropHelper-weighted raw copper count needs the unported item
// package for real Item construction (see Gravel's GetDropsForCompatibleTool doc comment for the
// same category of gap), so this returns nil for now.
func (c *CopperOre) GetDropsForCompatibleTool(item Item) []Item { return nil }
