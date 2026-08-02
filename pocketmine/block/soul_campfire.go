package block

// SoulCampfire is a port of pocketmine\block\SoulCampfire.
type SoulCampfire struct {
	Campfire
}

func NewSoulCampfire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SoulCampfire {
	s := &SoulCampfire{
		Campfire: Campfire{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
	}
	s.Init(s)
	return s
}

func (s *SoulCampfire) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SoulCampfire) GetLightLevel() int {
	if s.Lit {
		return 10
	}
	return 0
}

// GetDropsForCompatibleTool is a port of SoulCampfire::getDropsForCompatibleTool.
func (s *SoulCampfire) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaSoulSoil())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}

func (s *SoulCampfire) GetEntityCollisionDamage() int { return 2 }
