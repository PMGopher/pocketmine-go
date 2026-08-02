package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

const FrostedIceMaxAge = 3

// FrostedIce is a port of pocketmine\block\FrostedIce.
type FrostedIce struct {
	Ice
	AgeComponent
}

func NewFrostedIce(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *FrostedIce {
	f := &FrostedIce{
		Ice:          Ice{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(FrostedIceMaxAge),
	}
	f.Init(f)
	return f
}

func (f *FrostedIce) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FrostedIce) DescribeBlockOnlyState(w runtime.DataDescriber) { f.DescribeAge(w) }

func (f *FrostedIce) OnNearbyBlockChange() {
	if world, err := f.position.GetWorld(); err == nil {
		world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), rand.Intn(21)+20) // 20-40
	}
}

func (f *FrostedIce) OnRandomTick() {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	pos := f.position.AsVector3()

	if (!f.checkAdjacentBlocks(4) || rand.Intn(3) == 0) &&
		world.GetHighestAdjacentFullLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) >= 12-f.Age {
		if f.tryMelt() {
			for _, blk := range f.GetAllSides() {
				if other, ok := blk.(*FrostedIce); ok {
					other.tryMelt()
				}
			}
		}
	} else {
		world.ScheduleDelayedBlockUpdate(pos, rand.Intn(21)+20) // 20-40
	}
}

func (f *FrostedIce) OnScheduledUpdate() { f.OnRandomTick() }

func (f *FrostedIce) checkAdjacentBlocks(requirement int) bool {
	world, err := f.position.GetWorld()
	if err != nil {
		return false
	}
	pos := f.position.AsVector3()

	found := 0
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			if x == 0 && z == 0 {
				continue
			}
			if _, ok := world.GetBlockAt(pos.FloorX()+x, pos.FloorY(), pos.FloorZ()+z).(*FrostedIce); ok {
				found++
				if found >= requirement {
					return true
				}
			}
		}
	}
	return false
}

// tryMelt is a port of FrostedIce::tryMelt: updates the ice's age, or replaces it with
// VanillaWater() (destroying it, returning true) once fully aged.
func (f *FrostedIce) tryMelt() bool {
	world, err := f.position.GetWorld()
	if err != nil {
		return false
	}

	if f.Age >= FrostedIceMaxAge {
		Melt(f.self, VanillaWater())
		return true
	}

	f.Age++
	if err := world.SetBlock(f.position, f.self); err != nil {
		panic(err)
	}
	world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), rand.Intn(21)+20) // 20-40
	return false
}

func (f *FrostedIce) IsAffectedBySilkTouch() bool { return false }
