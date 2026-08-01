package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const FireMaxAge = 15

// Fire is a port of pocketmine\block\Fire.
type Fire struct {
	BaseFire
	AgeComponent
}

func NewFire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Fire {
	f := &Fire{
		BaseFire:     BaseFire{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}},
		AgeComponent: NewAgeComponent(FireMaxAge),
	}
	f.Init(f)
	return f
}

func (f *Fire) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Fire) DescribeBlockOnlyState(w runtime.DataDescriber) { f.DescribeAge(w) }

func (f *Fire) GetFireDamage() int { return 1 }

func (f *Fire) canBeSupportedBy(blk Behavior) bool {
	return blk.GetSupportType(math.Up) == blockutils.SupportTypeFull
}

func (f *Fire) hasAdjacentFlammableBlocks() bool {
	geo := f.self.(blockGeometry)
	for _, face := range math.AllFacing {
		if geo.GetSide(face, 1).IsFlammable() {
			return true
		}
	}
	return false
}

// OnNearbyBlockChange is a port of Fire::onNearbyBlockChange. Converting into SoulFire or Air
// both need the block registry (VanillaBlocks) to construct the replacement, so those two
// branches are skipped; the "still supported" branch (reschedule) is fully portable.
func (f *Fire) OnNearbyBlockChange() {
	down := f.self.(blockGeometry).GetSide(math.Down, 1)
	if soulFireCanBeSupportedBy(down) {
		// TODO: convert into SoulFire via the block registry.
		return
	}
	if !f.canBeSupportedBy(down) && !f.hasAdjacentFlammableBlocks() {
		// TODO: convert into Air via the block registry (or UseBreakOn as elsewhere).
		if world, err := f.position.GetWorld(); err == nil {
			world.UseBreakOn(f.position.AsVector3())
		}
		return
	}
	if world, err := f.position.GetWorld(); err == nil {
		world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), 30+rand.Intn(11))
	}
}

func (f *Fire) TicksRandomly() bool { return true }

// OnRandomTick is a port of Fire::onRandomTick, minus everything that needs the block registry
// (VanillaBlocks.AIR() as a replacement state), BlockEventHelper.Spread, and World.GetDifficulty/
// IsChunkLoaded/IsInWorld for burnBlocksAround/spreadFire - none ported yet, so aging and
// rescheduling are the only parts that actually run for now.
func (f *Fire) OnRandomTick() {
	if f.Age < FireMaxAge && rand.Intn(3) == 0 {
		f.Age++
		if world, err := f.position.GetWorld(); err == nil {
			if err := world.SetBlock(f.position, f.self); err != nil {
				panic(err)
			}
		}
	}

	if world, err := f.position.GetWorld(); err == nil {
		world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), 30+rand.Intn(11))
	}
}

func (f *Fire) OnScheduledUpdate() { f.self.OnRandomTick() }
