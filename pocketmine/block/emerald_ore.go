package block

import "math/rand"

// EmeraldOre is a port of pocketmine\block\EmeraldOre.
type EmeraldOre struct {
	Opaque
}

func NewEmeraldOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *EmeraldOre {
	e := &EmeraldOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	e.Init(e)
	return e
}

func (e *EmeraldOre) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool is a port of EmeraldOre::getDropsForCompatibleTool.
func (e *EmeraldOre) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("emerald", FortuneWeighted(item, 1, 1)))
}

func (e *EmeraldOre) IsAffectedBySilkTouch() bool { return true }

func (e *EmeraldOre) GetXpDropAmount() int { return rand.Intn(5) + 3 } // 3-7
