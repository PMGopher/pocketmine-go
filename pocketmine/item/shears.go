package item

import "pocketmine-go/pocketmine/block"

// Shears is a port of pocketmine\item\Shears..
type Shears struct {
	Tool
}

func NewShears(identifier ItemIdentifier, name string, enchantmentTags ...string) *Shears {
	s := &Shears{}
	s.Init(s, identifier, name)
	s.enchantmentTags = enchantmentTags
	return s
}

func (s *Shears) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Shears) GetMaxDurability() int { return 239 }

func (s *Shears) GetBlockToolType() block.ToolType { return block.ToolTypeShears }

func (s *Shears) GetBlockToolHarvestLevel() int { return 1 }

func (s *Shears) GetBaseMiningEfficiency() float64 { return 15 }
