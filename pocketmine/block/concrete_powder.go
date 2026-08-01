package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// ConcretePowder is a port of pocketmine\block\ConcretePowder.
type ConcretePowder struct {
	Opaque
	ColorComponent
	FallableComponent
}

func NewConcretePowder(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ConcretePowder {
	c := &ConcretePowder{
		Opaque:         Opaque{NewBlock(idInfo, name, typeInfo)},
		ColorComponent: NewColorComponent(),
	}
	c.Init(c)
	return c
}

func (c *ConcretePowder) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *ConcretePowder) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeColor(w) }

// getAdjacentWater is fully ported and ready to use once VanillaBlocks exists - see
// OnNearbyBlockChange/TickFalling below.
func (c *ConcretePowder) getAdjacentWater() (*Water, bool) {
	for _, f := range math.AllFacing {
		if f == math.Down {
			continue
		}
		if w, ok := c.GetSide(f, 1).(*Water); ok {
			return w, true
		}
	}
	return nil, false
}

// OnNearbyBlockChange should turn into VanillaBlocks.CONCRETE() (via BlockEventHelper.Form) when
// touching water, or start falling (FallableComponent's unported onNearbyBlockChange gap)
// otherwise — needs the unported block registry and BlockEventHelper, so this is a no-op for now.
func (c *ConcretePowder) OnNearbyBlockChange() {}

// TickFalling should return VanillaBlocks.CONCRETE().SetColor(...) when touching water — needs the
// unported block registry, so this always keeps falling unchanged for now.
func (c *ConcretePowder) TickFalling() (Behavior, bool) { return nil, false }
