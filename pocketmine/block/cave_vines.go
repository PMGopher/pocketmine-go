package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const CaveVinesMaxAge = 25

// CaveVines is a port of pocketmine\block\CaveVines.
type CaveVines struct {
	Flowable
	AgeComponent

	Berries bool
	Head    bool
}

func NewCaveVines(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CaveVines {
	c := &CaveVines{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(CaveVinesMaxAge),
	}
	c.Init(c)
	return c
}

func (c *CaveVines) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CaveVines) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeAge(w)
	w.Bool(&c.Berries)
	w.Bool(&c.Head)
}

func (c *CaveVines) HasBerries() bool { return c.Berries }

func (c *CaveVines) SetBerries(berries bool) { c.Berries = berries }

func (c *CaveVines) IsHead() bool { return c.Head }

func (c *CaveVines) SetHead(head bool) { c.Head = head }

func (c *CaveVines) CanClimb() bool { return true }

func (c *CaveVines) GetLightLevel() int {
	if c.Berries {
		return 14
	}
	return 0
}

func (c *CaveVines) canBeSupportedAt(blk Behavior) bool {
	geo := blk.(blockGeometry)
	support := geo.GetSide(math.Up, 1)
	supportGeo := support.(blockGeometry)
	return support.GetSupportType(math.Down) == blockutils.SupportTypeFull || supportGeo.HasSameTypeId(c.self)
}

func (c *CaveVines) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *CaveVines) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

func (c *CaveVines) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	c.Age = rand.Intn(CaveVinesMaxAge + 1)
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract's fertilizer-growth branch needs BlockEventHelper (unported); the berry-picking
// branch needs World.DropItem (not in the ported World interface) and real Item construction, so
// this is a no-op for now; it still returns false when there's nothing to pick and no fertilizer,
// matching the PHP original's final `return false;`.
func (c *CaveVines) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

// OnRandomTick's downward growth into a new CaveVines segment needs World.IsInWorld and the block
// registry (VanillaBlocks); only the "head" flag recompute is fully portable.
func (c *CaveVines) OnRandomTick() {
	head := true
	if down, ok := c.GetSide(math.Down, 1).(blockGeometry); ok {
		head = !down.HasSameTypeId(c.self)
	}
	if head != c.Head {
		c.Head = head
		if world, err := c.position.GetWorld(); err == nil {
			if err := world.SetBlock(c.position, c.self); err != nil {
				panic(err)
			}
		}
	}
}

func (c *CaveVines) TicksRandomly() bool { return true }

func (c *CaveVines) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (c *CaveVines) HasEntityCollision() bool { return true }

func (c *CaveVines) OnEntityInside(entity Entity) bool {
	entity.ResetFallDistance()
	return false
}

func (c *CaveVines) GetDropsForCompatibleTool(item Item) []Item {
	if c.Berries {
		return c.Block.GetDropsForCompatibleTool(item)
	}
	return nil
}

func (c *CaveVines) IsAffectedBySilkTouch() bool { return true }

func (c *CaveVines) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
