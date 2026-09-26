package block

// Snow is a port of pocketmine\block\Snow.
type Snow struct {
	Opaque
}

func NewSnow(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Snow {
	s := &Snow{Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Snow) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Snow) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool is a port of Snow::getDropsForCompatibleTool.
func (s *Snow) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("snowball", 4))
}
