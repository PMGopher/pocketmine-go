package item

import "pocketmine-go/pocketmine/block"

// Sword is a port of pocketmine\item\Sword. See Axe's doc comment for why
// OnDestroyBlock/OnAttackEntity aren't ported.
type Sword struct {
	TieredTool
}

func NewSword(identifier ItemIdentifier, name string, tier ToolTier) *Sword {
	s := &Sword{TieredTool: TieredTool{Tier: tier}}
	s.Init(s, identifier, name)
	return s
}

func (s *Sword) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Sword) GetBlockToolType() block.ToolType { return block.ToolTypeSword }

func (s *Sword) GetAttackPoints() int { return s.Tier.GetBaseAttackPoints() }

func (s *Sword) GetBlockToolHarvestLevel() int { return 1 }

// GetMiningEfficiency is a port of Sword::getMiningEfficiency: swords break any block 1.5x faster
// than an empty hand, on top of Tool's usual isCorrectTool gate (inlined here rather than reached
// via Tool.GetMiningEfficiency's self-dispatch, since Sword needs to scale its result too).
func (s *Sword) GetMiningEfficiency(isCorrectTool bool) float64 {
	base := 1.0
	if isCorrectTool {
		base = s.GetBaseMiningEfficiency()
	}
	return base * 1.5
}

func (s *Sword) GetBaseMiningEfficiency() float64 { return 10 }
