package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// stemShaper lets concrete Stem leaf types (MelonStem, PumpkinStem) report which block type they
// grow into, without needing an actual Block instance from the unported block registry
// (VanillaBlocks) - PHP's getPlant() returns a real Block just to compare hasSameTypeId/getTypeId
// against, so a type ID is all Stem's own logic actually needs. Same self-dispatch problem and
// solution shape as RailShaper/pressurePlateShaper/candleShaper.
type stemShaper interface {
	GetPlantTypeID() int
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

// OnRandomTick should grow via CropGrowthHelper.CanGrow and BlockEventHelper.Grow, then either
// advance age or - once at max age - sprout the plant (MELON/PUMPKIN) sideways onto an eligible
// neighbouring Air block. Needs CropGrowthHelper and BlockEventHelper, neither ported yet (same
// gap as Crops.OnRandomTick's doc comment), so this is a no-op for now.
func (s *Stem) OnRandomTick() {}

// GetDropsForCompatibleTool should return [s.AsItem().SetCount(FortuneDropHelper.Binomial(...))] —
// needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (s *Stem) GetDropsForCompatibleTool(item Item) []Item { return nil }
