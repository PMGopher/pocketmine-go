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

// GetDropsForCompatibleTool is a port of Podzol::getDropsForCompatibleTool.
func (p *Podzol) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaDirt())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
