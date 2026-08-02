package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// stemShaper lets concrete Stem leaf types (MelonStem, PumpkinStem) report which block type they
// grow into. GetPlantTypeID predates VanillaBlocks (a cheap type-ID-only comparison,
// still used by OnNearbyBlockChange); GetPlant returns the real grown block now that
// VanillaMelon/VanillaPumpkin exist, needed by OnRandomTick's sprout-sideways logic. Same
// self-dispatch shape as RailShaper/pressurePlateShaper/candleShaper.
type stemShaper interface {
	GetPlantTypeID() int
	GetPlant() Behavior
}

// Stem is a port of pocketmine\block\Stem. Like Crops (which it embeds), this isn't meant to be
// instantiated directly - a concrete leaf type (MelonStem, PumpkinStem) must embed it, implement
// Clone, and satisfy stemShaper.
type Stem struct {
	Crops

	Facing math.Facing
}

func (s *Stem) DescribeBlockOnlyState(w runtime.DataDescriber) {
	s.Crops.DescribeBlockOnlyState(w)
	w.FacingExcept(&s.Facing, math.Down)
}

func (s *Stem) GetFacing() math.Facing { return s.Facing }

// SetFacing panics if facing is Down, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (s *Stem) SetFacing(facing math.Facing) {
	if facing == math.Down {
		panic("DOWN is not a valid facing for this block")
	}
	s.Facing = facing
}

func (s *Stem) OnNearbyBlockChange() {
	plantTypeID := s.self.(stemShaper).GetPlantTypeID()
	if s.Facing != math.Up && s.self.(blockGeometry).GetSide(s.Facing, 1).GetTypeId() != plantTypeID {
		if world, err := s.position.GetWorld(); err == nil {
			s.Facing = math.Up
			if err := world.SetBlock(s.position, s.self); err != nil {
				panic(err)
			}
		}
	}
	s.Crops.OnNearbyBlockChange()
}

func (s *Stem) TicksRandomly() bool { return s.Age < CropsMaxAge || s.Facing == math.Up }

// OnRandomTick is a port of Stem::onRandomTick.
func (s *Stem) OnRandomTick() {
	if s.Facing != math.Up || !CropGrowthCanGrow(s.self) {
		return
	}

	if s.Age < CropsMaxAge {
		clone := s.self.Clone()
		clone.(Ageable).SetAge(s.Age + 1)
		Grow(s.self, clone, nil)
		return
	}

	grow := s.self.(stemShaper).GetPlant()
	for _, side := range math.HorizontalFacing {
		neighbor := s.self.(blockGeometry).GetSide(side, 1)
		if neighbor.(blockGeometry).HasSameTypeId(grow) {
			return
		}
	}

	facing := math.HorizontalFacing[rand.Intn(len(math.HorizontalFacing))]
	sideBlock := s.self.(blockGeometry).GetSide(facing, 1)
	if sideBlock.GetTypeId() != AIR {
		return
	}
	below := sideBlock.(blockGeometry).GetSide(math.Down, 1)
	if !below.(blockGeometry).HasTypeTag(BlockTypeTagsDirt) {
		return
	}
	if Grow(sideBlock, grow, nil) {
		world, err := s.position.GetWorld()
		if err != nil {
			return
		}
		s.Facing = facing
		_ = world.SetBlock(s.position, s.self)
	}
}

// GetDropsForCompatibleTool should return [s.AsItem().SetCount(FortuneDropHelper.Binomial(...))] —
// needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (s *Stem) GetDropsForCompatibleTool(item Item) []Item { return nil }
