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

// GetDropsForCompatibleTool is a port of SeaLantern::getDropsForCompatibleTool.
func (s *SeaLantern) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("prismarine_crystals", min(5, FortuneDiscrete(item, 2, 3))))
}
