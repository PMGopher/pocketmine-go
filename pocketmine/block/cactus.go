package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/entity"
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

// OnEntityInside is a port of Cactus::onEntityInside.
func (c *Cactus) OnEntityInside(e Entity) bool {
	ev := entity.NewEntityDamageByBlockEvent(c.self, e, entity.EntityDamageCauseContact, 1, nil)
	e.Attack(ev)
	return true
}

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

// OnRandomTick is a port of Cactus::onRandomTick. Unlike most SetBlock callers in this package,
// the final self-persist calls ignore their error (`_ = world.SetBlock(...)`) rather than bailing
// out, matching the PHP original's fire-and-forget `$world->setBlock($this->position, $this,
// update: false)` at every return path.
func (c *Cactus) OnRandomTick() {
	geo := c.self.(blockGeometry)
	up := geo.GetSide(math.Up, 1)
	if up.GetTypeId() != AIR {
		return
	}

	world, err := c.position.GetWorld()
	if err != nil {
		return
	}

	upPos := c.position.GetSide(math.Up, 1)
	if !world.IsInWorld(upPos.FloorX(), upPos.FloorY(), upPos.FloorZ()) {
		return
	}

	height := 1
	for height < CactusMaxHeight && geo.GetSide(math.Down, height).(blockGeometry).HasSameTypeId(c.self) {
		height++
	}

	if c.Age == 9 {
		canGrowFlower := true
		upGeo := up.(blockGeometry)
		for _, side := range math.HorizontalFacing {
			if upGeo.GetSide(side, 1).IsSolid() {
				canGrowFlower = false
				break
			}
		}

		if canGrowFlower {
			chance := 10
			if height >= CactusMaxHeight {
				chance = 25
			}
			if rand.Intn(100)+1 <= chance {
				if Grow(up, VanillaCactusFlower(), nil) {
					c.Age = 0
					_ = world.SetBlock(c.position, c.self)
				}
				return
			}
		}
	}

	if c.Age == CactusMaxAge {
		c.Age = 0
		if height < CactusMaxHeight {
			Grow(up, VanillaCactus(), nil)
		}
	} else {
		c.Age++
	}
	_ = world.SetBlock(c.position, c.self)
}
