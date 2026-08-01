package block

import (
	"fmt"
	"sort"

	"pocketmine-go/pocketmine/math"
)

// RailShaper is implemented by concrete rail types (Rail, StraightOnlyRail) to plug into the
// shared BaseRail connection logic below. PHP's BaseRail calls these abstract/overridable methods
// on $this from within its own private methods (tryReconnect, getConnectedDirections, etc.) - the
// same self-dispatch need as Behavior/Block, but scoped to just the rail subsystem, so it gets its
// own small interface rather than growing the main Behavior interface for every block type.
//
// Since there's no natural embeddable "BaseRail" struct with a self field to set up (rails don't
// share any state, only behavior), the shared logic lives here as free functions taking self
// RailShaper explicitly, called by each concrete rail type's own Place/OnPostPlace/OnNearbyBlockChange.
type RailShaper interface {
	Behavior
	GetCurrentShapeConnections() []int
	SetShapeFromConnections(connections []int) error
	GetPossibleConnectionDirectionsOneConstraint(constraint int) map[int]bool
}

// railDefaultPossibleConnectionDirectionsOneConstraint is a port of
// BaseRail::getPossibleConnectionDirectionsOneConstraint (the non-Rail-overridden default).
func railDefaultPossibleConnectionDirectionsOneConstraint(constraint int) map[int]bool {
	opposite := math.Opposite(math.Facing(constraint &^ RailFlagAscend))
	possible := map[int]bool{int(opposite): true}
	if constraint&RailFlagAscend == 0 {
		possible[int(opposite)|RailFlagAscend] = true
	}
	return possible
}

func railGetPossibleConnectionDirections(self RailShaper, constraints []int) map[int]bool {
	switch len(constraints) {
	case 0:
		possible := map[int]bool{int(math.North): true, int(math.South): true, int(math.West): true, int(math.East): true}
		keys := make([]int, 0, len(possible))
		for p := range possible {
			keys = append(keys, p)
		}
		for _, p := range keys {
			possible[p|RailFlagAscend] = true
		}
		return possible
	case 1:
		return self.GetPossibleConnectionDirectionsOneConstraint(constraints[0])
	case 2:
		return map[int]bool{}
	default:
		panic(fmt.Sprintf("Expected at most 2 constraints, got %d", len(constraints)))
	}
}

func railGetConnectedDirections(self RailShaper) []int {
	selfGeo := self.(blockGeometry)
	var connections []int
	for _, connection := range self.GetCurrentShapeConnections() {
		facing := math.Facing(connection &^ RailFlagAscend)
		other := selfGeo.GetSide(facing, 1)
		otherConnection := int(math.Opposite(facing))

		if connection&RailFlagAscend != 0 {
			other = other.(blockGeometry).GetSide(math.Up, 1)
		} else if _, ok := other.(RailShaper); !ok {
			other = other.(blockGeometry).GetSide(math.Down, 1)
			otherConnection |= RailFlagAscend
		}

		if otherRail, ok := other.(RailShaper); ok {
			for _, oc := range otherRail.GetCurrentShapeConnections() {
				if oc == otherConnection {
					connections = append(connections, connection)
					break
				}
			}
		}
	}
	return connections
}

func railSetConnections(self RailShaper, connections []int) error {
	if len(connections) == 1 {
		connections = []int{connections[0], int(math.Opposite(math.Facing(connections[0] &^ RailFlagAscend)))}
	} else if len(connections) != 2 {
		return fmt.Errorf("expected exactly 2 connections, got %d", len(connections))
	}
	return self.SetShapeFromConnections(connections)
}

// railTryReconnect is a port of BaseRail::tryReconnect.
//
// PHP iterates $possible (an associative array) in insertion order, taking the first candidate
// that works; Go map iteration order is randomized, so this sorts the candidate directions first
// to keep the algorithm deterministic (not necessarily identical tie-breaking to the original,
// but consistent run-to-run rather than randomized).
func railTryReconnect(self RailShaper) {
	position := self.GetPosition()
	world, err := position.GetWorld()
	if err != nil {
		return
	}

	thisConnections := railGetConnectedDirections(self)
	changed := false

	for {
		possible := railGetPossibleConnectionDirections(self, thisConnections)
		possibleSides := make([]int, 0, len(possible))
		for side := range possible {
			possibleSides = append(possibleSides, side)
		}
		sort.Ints(possibleSides)

		cont := false

		for _, thisSide := range possibleSides {
			otherSide := int(math.Opposite(math.Facing(thisSide &^ RailFlagAscend)))
			other := self.(blockGeometry).GetSide(math.Facing(thisSide&^RailFlagAscend), 1)

			if thisSide&RailFlagAscend != 0 {
				other = other.(blockGeometry).GetSide(math.Up, 1)
			} else if _, ok := other.(RailShaper); !ok {
				other = other.(blockGeometry).GetSide(math.Down, 1)
				otherSide |= RailFlagAscend
			}

			otherRail, ok := other.(RailShaper)
			if !ok {
				continue
			}
			otherConnections := railGetConnectedDirections(otherRail)
			if len(otherConnections) >= 2 {
				continue
			}

			otherPossible := railGetPossibleConnectionDirections(otherRail, otherConnections)
			if otherPossible[otherSide] {
				otherConnections = append(otherConnections, otherSide)
				if err := railSetConnections(otherRail, otherConnections); err != nil {
					panic(err)
				}
				if err := world.SetBlock(otherRail.GetPosition(), otherRail); err != nil {
					panic(err)
				}

				changed = true
				thisConnections = append(thisConnections, thisSide)
				cont = len(thisConnections) < 2
				break
			}
		}

		if !cont {
			break
		}
	}

	if changed {
		if err := railSetConnections(self, thisConnections); err != nil {
			panic(err)
		}
		if err := world.SetBlock(position, self); err != nil {
			panic(err)
		}
	}
}

func railOnPostPlace(self RailShaper) { railTryReconnect(self) }

func railOnNearbyBlockChange(self RailShaper) {
	position := self.GetPosition()
	world, err := position.GetWorld()
	if err != nil {
		return
	}
	selfGeo := self.(blockGeometry)

	if !selfGeo.GetAdjacentSupportType(math.Down).HasEdgeSupport() {
		world.UseBreakOn(position.AsVector3())
		return
	}

	for _, connection := range self.GetCurrentShapeConnections() {
		if connection&RailFlagAscend != 0 {
			facing := math.Facing(connection &^ RailFlagAscend)
			if !selfGeo.GetSide(facing, 1).GetSupportType(math.Up).HasEdgeSupport() {
				world.UseBreakOn(position.AsVector3())
				break
			}
		}
	}
}

// railCanPlace is a port of BaseRail::place's support check (the actual `parent::place(...)` call
// is left to each concrete rail type, since it needs to go through its own embedded Block).
func railCanPlace(blockReplace Behavior) bool {
	return blockReplace.(blockGeometry).GetAdjacentSupportType(math.Down).HasEdgeSupport()
}
