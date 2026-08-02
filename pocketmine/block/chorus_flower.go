package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
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

// scanStem is a port of ChorusFlower::scanStem. Note stemHeight only increments when the loop
// completes an iteration normally (no break) - the same as the PHP original's
// `$yOffset++, $stemHeight++` for-loop post-expression, which break skips entirely.
func (c *ChorusFlower) scanStem() (stemHeight int, endStoneBelow bool) {
	world, err := c.position.GetWorld()
	if err != nil {
		return 0, false
	}
	for yOffset := 0; yOffset < chorusFlowerMaxStemHeight; yOffset, stemHeight = yOffset+1, stemHeight+1 {
		downPos := c.position.GetSide(math.Down, yOffset+1)
		down := world.GetBlockAt(downPos.FloorX(), downPos.FloorY(), downPos.FloorZ())
		if down.GetTypeId() != CHORUS_PLANT {
			if down.GetTypeId() == END_STONE {
				endStoneBelow = true
			}
			break
		}
	}
	return stemHeight, endStoneBelow
}

// allHorizontalBlocksEmpty is a port of ChorusFlower::allHorizontalBlocksEmpty. except is nil for
// PHP's `null` (no facing excluded).
func (c *ChorusFlower) allHorizontalBlocksEmpty(world World, position math.Vector3, except *math.Facing) bool {
	for facing, sidePos := range position.SidesAroundAxis(math.AxisY, 1) {
		if except != nil && facing == *except {
			continue
		}
		if world.GetBlockAt(sidePos.FloorX(), sidePos.FloorY(), sidePos.FloorZ()).GetTypeId() != AIR {
			return false
		}
	}
	return true
}

// canGrowUpwards is a port of ChorusFlower::canGrowUpwards.
func (c *ChorusFlower) canGrowUpwards(stemHeight int, endStoneBelow bool) bool {
	world, err := c.position.GetWorld()
	if err != nil {
		return false
	}

	up := c.position.GetSide(math.Up, 1)
	upAbove := c.position.GetSide(math.Up, 2)
	if !world.IsInWorld(up.FloorX(), up.FloorY(), up.FloorZ()) ||
		world.GetBlockAt(up.FloorX(), up.FloorY(), up.FloorZ()).GetTypeId() != AIR ||
		(world.IsInWorld(upAbove.FloorX(), upAbove.FloorY(), upAbove.FloorZ()) &&
			world.GetBlockAt(upAbove.FloorX(), upAbove.FloorY(), upAbove.FloorZ()).GetTypeId() != AIR) {
		return false
	}

	if c.self.(blockGeometry).GetSide(math.Down, 1).GetTypeId() != AIR {
		if stemHeight >= chorusFlowerMaxStemHeight {
			return false
		}
		bound := 3
		if endStoneBelow {
			bound = 4
		}
		if stemHeight > 1 && stemHeight > rand.Intn(bound+1) { // chance decreases for each added block of chorus plant
			return false
		}
	}

	return c.allHorizontalBlocksEmpty(world, up.AsVector3(), nil)
}

// grow is a port of ChorusFlower::grow.
func (c *ChorusFlower) grow(facing math.Facing, ageChange int, tx *BlockTransactionImpl) *BlockTransactionImpl {
	if tx == nil {
		world, err := c.position.GetWorld()
		if err != nil {
			return nil
		}
		tx = NewBlockTransaction(world)
	}
	newAge := c.Age + ageChange
	if newAge > ChorusFlowerMaxAge {
		newAge = ChorusFlowerMaxAge
	}
	grown := c.self.Clone().(*ChorusFlower)
	grown.Age = newAge
	tx.AddBlock(c.position.GetSide(facing, 1), grown)
	return tx
}

// OnRandomTick is a port of ChorusFlower::onRandomTick.
func (c *ChorusFlower) OnRandomTick() {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	if c.Age >= ChorusFlowerMaxAge {
		return
	}

	var tx *BlockTransactionImpl

	stemHeight, endStoneBelow := c.scanStem()
	if c.canGrowUpwards(stemHeight, endStoneBelow) {
		tx = c.grow(math.Up, 0, tx)
	} else {
		bound := 3
		if endStoneBelow {
			bound = 4
		}
		maxAttempts := rand.Intn(bound + 1)
		visited := map[math.Facing]bool{}
		for attempts := 0; attempts < maxAttempts; attempts++ {
			facing := math.HorizontalFacing[rand.Intn(len(math.HorizontalFacing))]
			if visited[facing] {
				continue
			}
			visited[facing] = true

			sidePos := c.position.GetSide(facing, 1)
			downPos := sidePos.GetSide(math.Down, 1)
			opposite := math.Opposite(facing)
			if world.GetBlockAt(sidePos.FloorX(), sidePos.FloorY(), sidePos.FloorZ()).GetTypeId() == AIR &&
				world.GetBlockAt(downPos.FloorX(), downPos.FloorY(), downPos.FloorZ()).GetTypeId() == AIR &&
				c.allHorizontalBlocksEmpty(world, sidePos.AsVector3(), &opposite) {
				tx = c.grow(facing, 1, tx)
			}
		}
	}

	if tx != nil {
		tx.AddBlock(c.position, VanillaChorusPlant())
		ev := &StructureGrowEvent{Block: c.self, Transaction: tx, Player: nil}
		event.Call(ev)
		if !ev.IsCancelled() && tx.Apply() {
			world.AddSound(c.position.AsVector3().Add(0.5, 0.5, 0.5), sound.ChorusFlowerGrowSound{})
		}
	} else {
		world.AddSound(c.position.AsVector3().Add(0.5, 0.5, 0.5), sound.ChorusFlowerDieSound{})
		dead := c.self.Clone().(*ChorusFlower)
		dead.Age = ChorusFlowerMaxAge
		_ = world.SetBlock(c.position, dead)
	}
}
