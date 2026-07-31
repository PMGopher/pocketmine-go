package block

// Melon is a port of pocketmine\block\Melon.
type Melon struct {
	Opaque
}

func NewMelon(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Melon {
	m := &Melon{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *Melon) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool should return melon slices scaled via FortuneDropHelper — needs real
// Item construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (m *Melon) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (m *Melon) IsAffectedBySilkTouch() bool { return true }
