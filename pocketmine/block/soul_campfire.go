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

// GetDropsForCompatibleTool should return [VanillaBlocks.SOUL_SOIL().AsItem()] - needs the
// unported block registry (see Block.GetDropsForCompatibleTool's doc comment), so it's left as
// Block's default for now.

func (s *SoulCampfire) GetEntityCollisionDamage() int { return 2 }
