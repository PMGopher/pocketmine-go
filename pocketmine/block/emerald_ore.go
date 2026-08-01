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

// GetDropsForCompatibleTool should return an emerald scaled via FortuneDropHelper — needs real
// Item construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (e *EmeraldOre) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (e *EmeraldOre) IsAffectedBySilkTouch() bool { return true }

func (e *EmeraldOre) GetXpDropAmount() int { return rand.Intn(5) + 3 } // 3-7
