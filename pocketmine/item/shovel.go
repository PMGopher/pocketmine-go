package item

import "pocketmine-go/pocketmine/block"

// Shovel is a port of pocketmine\item\Shovel. See Axe's doc comment for why
// OnDestroyBlock/OnAttackEntity aren't ported.
type Shovel struct {
	TieredTool
}

func NewShovel(identifier ItemIdentifier, name string, tier ToolTier) *Shovel {
	s := &Shovel{TieredTool: TieredTool{Tier: tier}}
	s.Init(s, identifier, name)
	return s
}

func (s *Shovel) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Shovel) GetBlockToolType() block.ToolType { return block.ToolTypeShovel }

func (s *Shovel) GetBlockToolHarvestLevel() int { return s.Tier.GetHarvestLevel() }

func (s *Shovel) GetAttackPoints() int { return s.Tier.GetBaseAttackPoints() - 3 }
