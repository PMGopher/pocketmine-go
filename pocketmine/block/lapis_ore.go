package block

import "math/rand"

// LapisOre is a port of pocketmine\block\LapisOre.
type LapisOre struct {
	Opaque
}

func NewLapisOre(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *LapisOre {
	l := &LapisOre{Opaque{NewBlock(idInfo, name, typeInfo)}}
	l.Init(l)
	return l
}

func (l *LapisOre) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool should return lapis lazuli scaled via FortuneDropHelper — needs real
// Item construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (l *LapisOre) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (l *LapisOre) IsAffectedBySilkTouch() bool { return true }

func (l *LapisOre) GetXpDropAmount() int { return rand.Intn(4) + 2 } // 2-5
