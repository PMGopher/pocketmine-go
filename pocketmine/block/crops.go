package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const CropsMaxAge = 7

// Crops is a port of pocketmine\block\Crops.
//
// Like Button, this isn't meant to be instantiated directly - it has no Clone() of its own, so a
// concrete leaf type (Wheat, Carrot, Potato, Beetroot - not yet ported, all need the item
// registry for their AsItem()/GetDropsForCompatibleTool overrides) must embed it and implement
// Clone.
type Crops struct {
	Flowable
	AgeComponent
}

func (c *Crops) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeAge(w) }

func (c *Crops) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == FARMLAND
}

func (c *Crops) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *Crops) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract is a port of Crops::onInteract. `$item instanceof Fertilizer` is checked via item
// type ID (bone meal is the only Fertilizer-marked item in the PHP original) rather than a
// dedicated marker interface - same structural-marker convention as Durable/Axe/Shovel elsewhere
// in this port.
func (c *Crops) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() != itemTypeIDsBoneMeal {
		return false
	}
	newAge := c.Age + 2 + rand.Intn(4) // mt_rand(2, 5)
	if newAge > c.MaxAge {
		newAge = c.MaxAge
	}
	clone := c.self.Clone()
	clone.(Ageable).SetAge(newAge)
	if Grow(c.self, clone, player) {
		item.Pop()
	}
	return true
}

func (c *Crops) TicksRandomly() bool { return c.Age < CropsMaxAge }

// OnRandomTick is a port of Crops::onRandomTick.
func (c *Crops) OnRandomTick() {
	if c.Age < CropsMaxAge && CropGrowthCanGrow(c.self) {
		clone := c.self.Clone()
		clone.(Ageable).SetAge(c.Age + 1)
		Grow(c.self, clone, nil)
	}
}
