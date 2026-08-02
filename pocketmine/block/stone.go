package block

// Stone is a port of the anonymous Opaque subclass VanillaBlocksInputs.php registers for "stone".
type Stone struct {
	Opaque
}

func NewStone(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Stone {
	s := &Stone{Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Stone) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Stone) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool is a port of Stone's getDropsForCompatibleTool override.
func (s *Stone) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaCobblestone())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
