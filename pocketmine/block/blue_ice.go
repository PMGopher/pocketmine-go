package block

// BlueIce is a port of pocketmine\block\BlueIce.
type BlueIce struct {
	Opaque
}

func NewBlueIce(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BlueIce {
	b := &BlueIce{Opaque{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *BlueIce) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BlueIce) GetFrictionFactor() float64 { return 0.99 }

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (b *BlueIce) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (b *BlueIce) IsAffectedBySilkTouch() bool { return true }
