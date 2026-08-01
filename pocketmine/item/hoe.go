package item

import "pocketmine-go/pocketmine/block"

// Hoe is a port of pocketmine\item\Hoe. See Axe's doc comment for why
// OnDestroyBlock/OnAttackEntity aren't ported.
type Hoe struct {
	TieredTool
}

func NewHoe(identifier ItemIdentifier, name string, tier ToolTier) *Hoe {
	h := &Hoe{TieredTool: TieredTool{Tier: tier}}
	h.Init(h, identifier, name)
	return h
}

func (h *Hoe) Clone() Item {
	c := *h
	c.rebind(&c)
	return &c
}

func (h *Hoe) GetBlockToolType() block.ToolType { return block.ToolTypeHoe }
