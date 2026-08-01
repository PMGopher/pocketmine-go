package block

// SeaLantern is a port of pocketmine\block\SeaLantern.
type SeaLantern struct {
	Transparent
}

func NewSeaLantern(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SeaLantern {
	s := &SeaLantern{Transparent{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *SeaLantern) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SeaLantern) GetLightLevel() int { return 15 }

func (s *SeaLantern) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool's FortuneDropHelper-scaled prismarine crystal count needs the
// unported item package for real Item construction (see Gravel's GetDropsForCompatibleTool doc
// comment for the same category of gap), so this returns nil for now.
func (s *SeaLantern) GetDropsForCompatibleTool(item Item) []Item { return nil }
