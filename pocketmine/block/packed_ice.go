package block

// PackedIce is a port of pocketmine\block\PackedIce.
type PackedIce struct {
	Opaque
}

func NewPackedIce(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PackedIce {
	p := &PackedIce{Opaque{NewBlock(idInfo, name, typeInfo)}}
	p.Init(p)
	return p
}

func (p *PackedIce) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PackedIce) GetFrictionFactor() float64 { return 0.98 }

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (p *PackedIce) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (p *PackedIce) IsAffectedBySilkTouch() bool { return true }
