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

// GetDropsForCompatibleTool should return [VanillaItems.SNOWBALL().SetCount(4)] — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (s *Snow) GetDropsForCompatibleTool(item Item) []Item { return nil }
