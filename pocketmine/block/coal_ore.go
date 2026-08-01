package block

import "math/rand"

// CoalOre is a port of pocketmine\block\CoalOre.
type CoalOre struct {
	Opaque
}

func NewCoalOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CoalOre {
	c := &CoalOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CoalOre) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

// GetDropsForCompatibleTool should return coal scaled via FortuneDropHelper — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (c *CoalOre) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (c *CoalOre) IsAffectedBySilkTouch() bool { return true }

func (c *CoalOre) GetXpDropAmount() int { return rand.Intn(3) } // 0-2
