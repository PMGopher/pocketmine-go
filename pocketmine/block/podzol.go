package block

// Podzol is a port of pocketmine\block\Podzol.
type Podzol struct {
	Opaque
}

func NewPodzol(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Podzol {
	p := &Podzol{Opaque{NewBlock(idInfo, name, typeInfo)}}
	p.Init(p)
	return p
}

func (p *Podzol) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Podzol) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool should return VanillaBlocks.DIRT().AsItem() — needs the unported
// block registry (VanillaBlocks) and real Item construction (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (p *Podzol) GetDropsForCompatibleTool(item Item) []Item { return nil }
