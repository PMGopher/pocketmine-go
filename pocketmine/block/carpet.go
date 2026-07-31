package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Carpet is a port of pocketmine\block\Carpet.
type Carpet struct {
	Flowable
	ColorComponent
}

func NewCarpet(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Carpet {
	c := &Carpet{
		Flowable:       Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		ColorComponent: NewColorComponent(),
	}
	c.Init(c)
	return c
}

func (c *Carpet) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Carpet) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeColor(w) }

func (c *Carpet) IsSolid() bool { return true }

func (c *Carpet) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 15.0/16.0)}
}

func (c *Carpet) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() != AIR
}

func (c *Carpet) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *Carpet) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

func (c *Carpet) GetFlameEncouragement() int { return 30 }

func (c *Carpet) GetFlammability() int { return 20 }
