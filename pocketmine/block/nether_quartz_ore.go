package block

import "math/rand"

// NetherQuartzOre is a port of pocketmine\block\NetherQuartzOre.
type NetherQuartzOre struct {
	Opaque
}

func NewNetherQuartzOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherQuartzOre {
	n := &NetherQuartzOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	n.Init(n)
	return n
}

func (n *NetherQuartzOre) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherQuartzOre) IsAffectedBySilkTouch() bool { return true }

func (n *NetherQuartzOre) GetXpDropAmount() int { return rand.Intn(4) + 2 } // 2-5

// GetDropsForCompatibleTool's FortuneDropHelper-weighted nether quartz count needs the unported
// item package for real Item construction (see Gravel's GetDropsForCompatibleTool doc comment for
// the same category of gap), so this returns nil for now.
func (n *NetherQuartzOre) GetDropsForCompatibleTool(item Item) []Item { return nil }
