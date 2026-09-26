package item

import "pocketmine-go/pocketmine/block"

// Axe is a port of pocketmine\item\Axe. GetBlockToolType()==block.ToolTypeAxe and its embedded
// Durable.ApplyDamage together mean a real Axe now satisfies both block.Axe (wood.go) and
// block.Durable (candle_component.go).

type Axe struct {
	TieredTool
}

func NewAxe(identifier ItemIdentifier, name string, tier ToolTier, enchantmentTags ...string) *Axe {
	a := &Axe{TieredTool: TieredTool{Tier: tier}}
	a.Init(a, identifier, name)
	a.enchantmentTags = enchantmentTags
	return a
}

func (a *Axe) Clone() Item {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *Axe) GetBlockToolType() block.ToolType { return block.ToolTypeAxe }

func (a *Axe) GetBlockToolHarvestLevel() int { return a.Tier.GetHarvestLevel() }

func (a *Axe) GetAttackPoints() int { return a.Tier.GetBaseAttackPoints() - 1 }
