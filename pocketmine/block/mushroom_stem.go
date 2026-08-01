package block

// MushroomStem is a port of pocketmine\block\MushroomStem.
type MushroomStem struct {
	Opaque
}

func NewMushroomStem(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *MushroomStem {
	m := &MushroomStem{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *MushroomStem) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MushroomStem) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (m *MushroomStem) IsAffectedBySilkTouch() bool { return true }
