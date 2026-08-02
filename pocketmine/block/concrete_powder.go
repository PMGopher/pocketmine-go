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

// OnNearbyBlockChange is a port of ConcretePowder::onNearbyBlockChange, minus the "else start
// falling" branch (FallableTrait::onNearbyBlockChange needs the unported FallingBlock entity type
// - see FallableComponent's doc comment), so nothing happens when there's no adjacent water yet.
func (c *ConcretePowder) OnNearbyBlockChange() {
	if water, ok := c.getAdjacentWater(); ok {
		Form(c.self, c.concreteOfMyColor(), water)
	}
}

// TickFalling is a port of ConcretePowder::tickFalling.
func (c *ConcretePowder) TickFalling() (Behavior, bool) {
	if _, ok := c.getAdjacentWater(); !ok {
		return nil, false
	}
	return c.concreteOfMyColor(), true
}

func (c *ConcretePowder) concreteOfMyColor() Behavior {
	concrete := VanillaConcrete().(*Concrete)
	concrete.SetColor(c.Color)
	return concrete
}
