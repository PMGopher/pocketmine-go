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

// GetDropsForCompatibleTool is a port of Melon::getDropsForCompatibleTool.
func (m *Melon) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("melon", min(9, FortuneDiscrete(item, 3, 7))))
}

func (m *Melon) IsAffectedBySilkTouch() bool { return true }
