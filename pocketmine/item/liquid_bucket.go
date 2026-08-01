package item

import "pocketmine-go/pocketmine/block"

// LiquidBucket is a port of pocketmine\item\LiquidBucket. OnInteractBlock (emptying onto a
// replaceable block) needs a real Player/Block/World - see the Item interface's doc comment.
// GetFuelResidue (VanillaItems.BUCKET()) isn't ported either - needs the item registry.
type LiquidBucket struct {
	ItemBase

	Liquid block.Behavior
}

func NewLiquidBucket(identifier ItemIdentifier, name string, liquid block.Behavior) *LiquidBucket {
	l := &LiquidBucket{Liquid: liquid}
	l.Init(l, identifier, name)
	return l
}

func (l *LiquidBucket) Clone() Item {
	c := *l
	c.Liquid = l.Liquid.Clone()
	c.rebind(&c)
	return &c
}

func (l *LiquidBucket) GetMaxStackSize() int { return 1 }

// GetFuelTime is a port of LiquidBucket::getFuelTime.
func (l *LiquidBucket) GetFuelTime() int {
	if _, ok := l.Liquid.(*block.Lava); ok {
		return 20000
	}
	return 0
}

func (l *LiquidBucket) GetLiquid() block.Behavior { return l.Liquid }
