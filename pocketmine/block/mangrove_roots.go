package block

// MangroveRoots is a port of pocketmine\block\MangroveRoots.
type MangroveRoots struct {
	Transparent
}

func NewMangroveRoots(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *MangroveRoots {
	m := &MangroveRoots{Transparent{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *MangroveRoots) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MangroveRoots) GetFlammability() int { return 5 }

func (m *MangroveRoots) GetFlameEncouragement() int { return 5 }
