package block

import (
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

// OnInteract should fertilize the crop (advancing its age by a random 2-5, via
// BlockEventHelper.Grow) — needs a Fertilizer item marker and BlockEventHelper, neither ported
// yet, so this is a no-op for now.
func (c *Crops) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

func (c *Crops) TicksRandomly() bool { return c.Age < CropsMaxAge }

// OnRandomTick should grow via CropGrowthHelper.CanGrow (light level + nearby-farmland-hydration
// multiplier) and BlockEventHelper.Grow — neither ported yet (CropGrowthHelper needs
// World.GetPotentialLightAt, not in the ported World interface either), so this is a no-op for now.
func (c *Crops) OnRandomTick() {}
