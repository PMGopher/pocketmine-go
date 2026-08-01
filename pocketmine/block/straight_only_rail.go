package block

import (
	"fmt"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// StraightOnlyRail is a port of pocketmine\block\StraightOnlyRail: a simple non-curvable rail.
//
// Like Button/Crops, this isn't meant to be instantiated directly - it has no Clone() of its own.
// Its concrete subtypes (DetectorRail, ActivatorRail, PoweredRail - not yet ported, all need
// RailPoweredByRedstoneTrait for their extra "powered" state) must embed it and implement Clone.
type StraightOnlyRail struct {
	Flowable

	RailShapeValue int
}

func (s *StraightOnlyRail) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Int(3, &s.RailShapeValue)
}

func (s *StraightOnlyRail) GetShape() int { return s.RailShapeValue }

// SetShape panics if shape isn't a valid straight rail shape, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (s *StraightOnlyRail) SetShape(shape int) {
	if _, ok := railConnections[shape]; !ok {
		panic(fmt.Sprintf("Invalid rail shape %d", shape))
	}
	s.RailShapeValue = shape
}

func (s *StraightOnlyRail) SetShapeFromConnections(connections []int) error {
	if shape, ok := railSearchState(connections, railConnections); ok {
		s.RailShapeValue = shape
		return nil
	}
	return fmt.Errorf("no rail shape matches these connections")
}

func (s *StraightOnlyRail) GetCurrentShapeConnections() []int {
	conn := railConnections[s.RailShapeValue]
	return []int{conn[0], conn[1]}
}

func (s *StraightOnlyRail) GetPossibleConnectionDirectionsOneConstraint(constraint int) map[int]bool {
	return railDefaultPossibleConnectionDirectionsOneConstraint(constraint)
}

func (s *StraightOnlyRail) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if railCanPlace(blockReplace) {
		return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

// OnPostPlace/OnNearbyBlockChange are promoted as-is by concrete subtypes once they call Init -
// self.(RailShaper) inside railOnPostPlace/railOnNearbyBlockChange then resolves to the concrete
// type via the usual Block.self mechanism.
func (s *StraightOnlyRail) OnPostPlace() { railOnPostPlace(s.self.(RailShaper)) }

func (s *StraightOnlyRail) OnNearbyBlockChange() { railOnNearbyBlockChange(s.self.(RailShaper)) }
