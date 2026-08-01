package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Sapling is a port of pocketmine\block\Sapling.
type Sapling struct {
	Flowable

	Ready       bool
	SaplingType blockutils.SaplingType
}

func NewSapling(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, saplingType blockutils.SaplingType) *Sapling {
	s := &Sapling{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, SaplingType: saplingType}
	s.Init(s)
	return s
}

func (s *Sapling) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Sapling) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&s.Ready) }

func (s *Sapling) IsReady() bool { return s.Ready }

func (s *Sapling) SetReady(ready bool) { s.Ready = ready }

func (s *Sapling) GetSaplingType() blockutils.SaplingType { return s.SaplingType }

func (s *Sapling) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud)
}

func (s *Sapling) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return s.canBeSupportedAt(blockReplace) && s.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (s *Sapling) OnNearbyBlockChange() {
	if !s.canBeSupportedAt(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker (pocketmine\item\Fertilizer,
// not ported yet - same gap documented on Crops/SweetBerryBush/CocoaBlock's OnInteract). Block's
// default OnInteract (return false) already matches this gap, so there's nothing to override
// here.

func (s *Sapling) TicksRandomly() bool { return true }

func (s *Sapling) OnRandomTick() {
	world, err := s.position.GetWorld()
	if err != nil {
		return
	}
	pos := s.position.AsVector3()
	if world.GetFullLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) < 8 || rand.Intn(7) != 0 {
		return
	}
	if s.Ready {
		s.grow()
	} else {
		s.Ready = true
		if err := world.SetBlock(s.position, s.self); err != nil {
			panic(err)
		}
	}
}

// grow is a port of Sapling::grow. It needs a Random-seeded TreeFactory (world-gen, not ported)
// and StructureGrowEvent (event package, not ported), so this is a no-op stub returning false for
// now - see Sugarcane.grow's doc comment for the same category of gap.
func (s *Sapling) grow() bool { return false }

func (s *Sapling) GetFuelTime() int { return 100 }
