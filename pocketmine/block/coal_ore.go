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

// GetDropsForCompatibleTool is a port of CoalOre::getDropsForCompatibleTool.
func (c *CoalOre) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("coal", FortuneWeighted(item, 1, 1)))
}

func (c *CoalOre) IsAffectedBySilkTouch() bool { return true }

func (c *CoalOre) GetXpDropAmount() int { return rand.Intn(3) } // 0-2
