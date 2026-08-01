package item

import "pocketmine-go/pocketmine/block"

// Pickaxe is a port of pocketmine\item\Pickaxe. See Axe's doc comment for why
// OnDestroyBlock/OnAttackEntity aren't ported.
type Pickaxe struct {
	TieredTool
}

func NewPickaxe(identifier ItemIdentifier, name string, tier ToolTier) *Pickaxe {
	p := &Pickaxe{TieredTool: TieredTool{Tier: tier}}
	p.Init(p, identifier, name)
	return p
}

func (p *Pickaxe) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Pickaxe) GetBlockToolType() block.ToolType { return block.ToolTypePickaxe }

func (p *Pickaxe) GetBlockToolHarvestLevel() int { return p.Tier.GetHarvestLevel() }

func (p *Pickaxe) GetAttackPoints() int { return p.Tier.GetBaseAttackPoints() - 2 }
