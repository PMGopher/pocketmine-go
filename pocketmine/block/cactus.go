package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	CactusMaxAge    = 15
	CactusMaxHeight = 3
)

// Cactus is a port of pocketmine\block\Cactus.
type Cactus struct {
	Transparent
	AgeComponent
}

func NewCactus(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Cactus {
	c := &Cactus{
		Transparent:  Transparent{NewBlock(idInfo, name, typeInfo)},
		AgeComponent: NewAgeComponent(CactusMaxAge),
	}
	c.Init(c)
	return c
}

func (c *Cactus) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Cactus) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeAge(w) }

func (c *Cactus) HasEntityCollision() bool { return true }

func (c *Cactus) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	const shrinkSize = 1.0 / 16
	return []math.AxisAlignedBB{math.OneAABB().ContractedCopy(shrinkSize, 0, shrinkSize).TrimmedCopy(math.Up, shrinkSize)}
}

func (c *Cactus) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// OnEntityInside should damage the entity via EntityDamageByBlockEvent - needs that unported
// concrete event subclass and Entity.Attack, so this is a no-op for now; it still returns true,
// matching the PHP original's unconditional `return true;`.
func (c *Cactus) OnEntityInside(entity Entity) bool { return true }

func (c *Cactus) canBeSupportedAt(blk Behavior) bool {
	geo := blk.(blockGeometry)
	support := geo.GetSide(math.Down, 1)
	supportGeo := support.(blockGeometry)
	if !supportGeo.HasSameTypeId(c.self) && !supportGeo.HasTypeTag(BlockTypeTagsSand) {
		return false
	}
	for _, side := range math.HorizontalFacing {
		if geo.GetSide(side, 1).IsSolid() {
			return false
		}
	}
	return true
}

func (c *Cactus) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *Cactus) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	}
}

func (c *Cactus) TicksRandomly() bool { return true }

// OnRandomTick should grow the cactus (or occasionally a CactusFlower) upward - needs
// World.IsInWorld, BlockEventHelper, and the block registry (VanillaBlocks), none ported yet, so
// this is a no-op for now.
func (c *Cactus) OnRandomTick() {}
