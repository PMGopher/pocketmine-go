package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const CocoaBlockMaxAge = 2

// CocoaBlock is a port of pocketmine\block\CocoaBlock.
type CocoaBlock struct {
	Flowable
	HorizontalFacingComponent
	AgeComponent
}

func NewCocoaBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CocoaBlock {
	c := &CocoaBlock{
		Flowable:                  Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		AgeComponent:              NewAgeComponent(CocoaBlockMaxAge),
	}
	c.Init(c)
	return c
}

func (c *CocoaBlock) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CocoaBlock) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
	age := c.Age
	w.BoundedIntAuto(0, CocoaBlockMaxAge, &age)
	c.Age = age
}

func (c *CocoaBlock) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	age := float64(c.Age)
	bb := math.OneAABB().
		SquashedCopy(math.FacingAxis(math.RotateY(c.Facing, true)), (6-age)/16).
		TrimmedCopy(math.Down, (7-age*2)/16).
		TrimmedCopy(math.Up, 0.25).
		TrimmedCopy(math.Opposite(c.Facing), 1.0/16).
		TrimmedCopy(c.Facing, (11-age*2)/16)
	return []math.AxisAlignedBB{bb}
}

// canAttachTo is a port of CocoaBlock::canAttachTo. PHP checks `instanceof Wood` specifically
// (the log block), not any wood-material block, so this asserts the concrete *Wood type rather
// than the broader WoodMaterial interface.
func (c *CocoaBlock) canAttachTo(blk Behavior) bool {
	wood, ok := blk.(*Wood)
	return ok && wood.GetWoodType() == blockutils.WoodTypeJungle
}

func (c *CocoaBlock) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if math.FacingAxis(face) != math.AxisY && c.canAttachTo(blockClicked) {
		c.Facing = face
		return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker and BlockEventHelper,
// neither ported yet. Block's default OnInteract (return false) already matches this gap, so
// there's nothing to override here.

func (c *CocoaBlock) OnNearbyBlockChange() {
	if !c.canAttachTo(c.self.(blockGeometry).GetSide(math.Opposite(c.Facing), 1)) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	}
}

func (c *CocoaBlock) TicksRandomly() bool { return c.Age < CocoaBlockMaxAge }

// OnRandomTick is a port of CocoaBlock::onRandomTick.
func (c *CocoaBlock) OnRandomTick() {
	if rand.Intn(5) == 0 { // mt_rand(1, 5) === 1
		c.grow(nil)
	}
}

// grow is a port of CocoaBlock::grow.
func (c *CocoaBlock) grow(player Player) bool {
	if c.Age < CocoaBlockMaxAge {
		next := c.self.Clone().(*CocoaBlock)
		next.Age++
		return Grow(c.self, next, player)
	}
	return false
}

// GetDropsForCompatibleTool/AsItem should return VanillaItems.COCOA_BEANS() (scaled when mature)
// — needs the unported item package (see Block.GetDropsForCompatibleTool's doc comment), so both
// are left as Block's defaults for now.
