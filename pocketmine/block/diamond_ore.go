package block

import "math/rand"

// DiamondOre is a port of pocketmine\block\DiamondOre.
type DiamondOre struct {
	Opaque
}

func NewDiamondOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DiamondOre {
	d := &DiamondOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	d.Init(d)
	return d
}

func (d *DiamondOre) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool is a port of DiamondOre::getDropsForCompatibleTool.
func (d *DiamondOre) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("diamond", FortuneWeighted(item, 1, 1)))
}

func (d *DiamondOre) IsAffectedBySilkTouch() bool { return true }

func (d *DiamondOre) GetXpDropAmount() int { return rand.Intn(5) + 3 } // 3-7
