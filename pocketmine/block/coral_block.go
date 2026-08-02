package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CoralBlock is a port of pocketmine\block\CoralBlock.
type CoralBlock struct {
	Opaque
	CoralComponent
}

func NewCoralBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CoralBlock {
	c := &CoralBlock{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CoralBlock) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CoralBlock) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCoral(w) }

func (c *CoralBlock) OnNearbyBlockChange() {
	if !c.Dead {
		if world, err := c.position.GetWorld(); err == nil {
			world.ScheduleDelayedBlockUpdate(c.position.AsVector3(), 40+rand.Intn(161))
		}
	}
}

// OnScheduledUpdate is a port of CoralBlock::onScheduledUpdate.
func (c *CoralBlock) OnScheduledUpdate() {
	if c.Dead {
		return
	}
	geo := c.self.(blockGeometry)
	hasWater := false
	for _, face := range math.AllFacing {
		if geo.GetSide(face, 1).GetTypeId() == WATER {
			hasWater = true
			break
		}
	}
	if !hasWater {
		deadClone := c.self.Clone()
		deadClone.(CoralMaterial).SetDead(true)
		Die(c.self, deadClone)
	}
}

func (c *CoralBlock) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool is a port of CoralBlock::getDropsForCompatibleTool - coral always
// drops its dead form, regardless of whether the block itself is alive or dead when broken.
func (c *CoralBlock) GetDropsForCompatibleTool(item Item) []Item {
	dead := c.self.Clone()
	dead.(CoralMaterial).SetDead(true)
	dropped := asItemOrNil(dead)
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
