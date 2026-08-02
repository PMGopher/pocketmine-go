package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
)

const (
	FarmlandMaxWetness = 7

	farmlandWaterSearchHorizontalLength = 9
	farmlandWaterSearchVerticalLength   = 2
	farmlandWaterPositionIndexUnknown   = -1
	farmlandWaterPositionIndicesTotal   = farmlandWaterSearchHorizontalLength * farmlandWaterSearchHorizontalLength * farmlandWaterSearchVerticalLength
)

// Farmland is a port of pocketmine\block\Farmland.
type Farmland struct {
	Transparent

	Wetness            int // "moisture" blockstate property in PC
	WaterPositionIndex int // internal cache only, not exposed on the wire - see PHP doc comment
}

func NewFarmland(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Farmland {
	f := &Farmland{
		Transparent:        Transparent{NewBlock(idInfo, name, typeInfo)},
		WaterPositionIndex: farmlandWaterPositionIndexUnknown,
	}
	f.Init(f)
	return f
}

func (f *Farmland) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Farmland) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.BoundedIntAuto(0, FarmlandMaxWetness, &f.Wetness)
	w.BoundedIntAuto(-1, farmlandWaterPositionIndicesTotal-1, &f.WaterPositionIndex)
}

func (f *Farmland) GetWetness() int { return f.Wetness }

// SetWetness panics if wetness is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (f *Farmland) SetWetness(wetness int) {
	if wetness < 0 || wetness > FarmlandMaxWetness {
		panic("Wetness must be in range 0 ... 7")
	}
	f.Wetness = wetness
}

func (f *Farmland) GetWaterPositionIndex() int { return f.WaterPositionIndex }

func (f *Farmland) SetWaterPositionIndex(index int) {
	if index < -1 || index >= farmlandWaterPositionIndicesTotal {
		panic("Water XZ index must be in range -1 ... 161")
	}
	f.WaterPositionIndex = index
}

func (f *Farmland) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 1.0/16)}
}

// OnNearbyBlockChange is a port of Farmland::onNearbyBlockChange.
func (f *Farmland) OnNearbyBlockChange() {
	if f.self.(blockGeometry).GetSide(math.Up, 1).IsSolid() {
		world, err := f.position.GetWorld()
		if err != nil {
			return
		}
		_ = world.SetBlock(f.position, VanillaDirt())
	}
}

func (f *Farmland) TicksRandomly() bool { return true }

// OnRandomTick doesn't fire FarmlandHydrationChangeEvent (deferred concrete event subclass), so
// hydration changes are never cancellable yet; the "dry out completely -> become Dirt" step is
// otherwise real.
func (f *Farmland) OnRandomTick() {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}

	oldWaterPositionIndex := f.WaterPositionIndex
	changed := false

	if !f.canHydrate() {
		if f.Wetness > 0 {
			f.Wetness--
			if err := world.SetBlock(f.position, f.self); err != nil {
				panic(err)
			}
			changed = true
		} else {
			_ = world.SetBlock(f.position, VanillaDirt())
			changed = true
		}
	} else if f.Wetness < FarmlandMaxWetness {
		f.Wetness = FarmlandMaxWetness
		if err := world.SetBlock(f.position, f.self); err != nil {
			panic(err)
		}
		changed = true
	}

	if !changed && oldWaterPositionIndex != f.WaterPositionIndex {
		if err := world.SetBlock(f.position, f.self); err != nil {
			panic(err)
		}
	}
}

// OnEntityLand is a port of Farmland::onEntityLand. It always defers to the default fall-damage
// behavior either way, matching the PHP original's `return null;` outcome.
func (f *Farmland) OnEntityLand(e Entity) (float64, bool) {
	if living, ok := e.(Living); ok && utils.GetRandomFloat() < living.GetFallDistance()-0.5 {
		ev := entity.NewEntityTrampleFarmlandEvent(living, f.self)
		entity.Call(ev)
		if !ev.IsCancelled() {
			if world, err := f.position.GetWorld(); err == nil {
				_ = world.SetBlock(f.position, VanillaDirt())
			}
		}
	}
	return 0, false
}

func (f *Farmland) canHydrate() bool {
	world, err := f.position.GetWorld()
	if err != nil {
		return false
	}

	startX := f.position.FloorX() - farmlandWaterSearchHorizontalLength/2
	startY := f.position.FloorY()
	startZ := f.position.FloorZ() - farmlandWaterSearchHorizontalLength/2

	if f.WaterPositionIndex != farmlandWaterPositionIndexUnknown {
		raw := f.WaterPositionIndex
		x := raw % farmlandWaterSearchHorizontalLength
		raw /= farmlandWaterSearchHorizontalLength
		z := raw % farmlandWaterSearchHorizontalLength
		raw /= farmlandWaterSearchHorizontalLength
		y := raw % farmlandWaterSearchVerticalLength

		if _, ok := world.GetBlockAt(startX+x, startY+y, startZ+z).(*Water); ok {
			return true
		}
	}

	// No water found at the cached position - search the whole area. y increments after x/z are
	// exhausted, since water is usually at the same Y as the farmland.
	for y := 0; y < farmlandWaterSearchVerticalLength; y++ {
		for x := 0; x < farmlandWaterSearchHorizontalLength; x++ {
			for z := 0; z < farmlandWaterSearchHorizontalLength; z++ {
				if _, ok := world.GetBlockAt(startX+x, startY+y, startZ+z).(*Water); ok {
					f.WaterPositionIndex = x + z*farmlandWaterSearchHorizontalLength + y*farmlandWaterSearchHorizontalLength*farmlandWaterSearchHorizontalLength
					return true
				}
			}
		}
	}

	f.WaterPositionIndex = farmlandWaterPositionIndexUnknown
	return false
}

// GetDropsForCompatibleTool/GetPickedItem should return VanillaBlocks.DIRT().AsItem() — needs the
// unported block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc
// comment), so these return nil for now.
func (f *Farmland) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (f *Farmland) GetPickedItem(addUserData bool) Item { return nil }
