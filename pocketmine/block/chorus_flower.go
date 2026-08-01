package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	ChorusFlowerMinAge        = 0
	ChorusFlowerMaxAge        = 5
	chorusFlowerMaxStemHeight = 5
)

// ChorusFlower is a port of pocketmine\block\ChorusFlower.
type ChorusFlower struct {
	Flowable
	AgeComponent
}

func NewChorusFlower(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ChorusFlower {
	c := &ChorusFlower{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(ChorusFlowerMaxAge),
	}
	c.Init(c)
	return c
}

func (c *ChorusFlower) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *ChorusFlower) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeAge(w) }

func (c *ChorusFlower) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB()}
}

func (c *ChorusFlower) canBeSupportedAt(blk Behavior) bool {
	geo := blk.(blockGeometry)
	world, err := blk.GetPosition().GetWorld()
	if err != nil {
		return false
	}

	down := geo.GetSide(math.Down, 1)
	if down.GetTypeId() == END_STONE || down.GetTypeId() == CHORUS_PLANT {
		return true
	}

	plantAdjacent := false
	for _, sidePos := range blk.GetPosition().AsVector3().SidesAroundAxis(math.AxisY, 1) {
		side := world.GetBlockAt(sidePos.FloorX(), sidePos.FloorY(), sidePos.FloorZ())
		switch {
		case side.GetTypeId() == CHORUS_PLANT:
			if plantAdjacent { // at most one plant may be horizontally adjacent
				return false
			}
			plantAdjacent = true
		case side.GetTypeId() != AIR:
			return false
		}
	}

	return plantAdjacent
}

func (c *ChorusFlower) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *ChorusFlower) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

func (c *ChorusFlower) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	if world, err := c.position.GetWorld(); err == nil {
		world.UseBreakOn(c.position.AsVector3())
	}
}

func (c *ChorusFlower) TicksRandomly() bool { return c.Age < ChorusFlowerMaxAge }

// OnRandomTick should grow the flower upward or sideways into a chorus plant stem (a whole
// branching-growth algorithm: scanStem/canGrowUpwards/allHorizontalBlocksEmpty/grow) - needs
// World.IsInWorld, the block registry (VanillaBlocks.CHORUS_PLANT()), BlockTransaction
// construction, and StructureGrowEvent, none ported yet, so this is a no-op for now.
func (c *ChorusFlower) OnRandomTick() {}
