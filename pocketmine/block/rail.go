package block

import (
	"fmt"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Rail is a port of pocketmine\block\Rail.
type Rail struct {
	Flowable

	RailShapeValue int
}

func NewRail(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Rail {
	r := &Rail{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, RailShapeValue: RailStraightNorthSouth}
	r.Init(r)
	return r
}

func (r *Rail) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *Rail) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Int(4, &r.RailShapeValue) }

func (r *Rail) GetShape() int { return r.RailShapeValue }

// SetShape panics if shape isn't a valid rail or curve shape, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (r *Rail) SetShape(shape int) {
	if _, ok := railConnections[shape]; !ok {
		if _, ok := railCurveConnections[shape]; !ok {
			panic(fmt.Sprintf("Invalid shape %d", shape))
		}
	}
	r.RailShapeValue = shape
}

func (r *Rail) SetShapeFromConnections(connections []int) error {
	if shape, ok := railSearchState(connections, railConnections); ok {
		r.RailShapeValue = shape
		return nil
	}
	if shape, ok := railSearchState(connections, railCurveConnections); ok {
		r.RailShapeValue = shape
		return nil
	}
	return fmt.Errorf("no rail shape matches these connections")
}

func (r *Rail) GetCurrentShapeConnections() []int {
	if conn, ok := railCurveConnections[r.RailShapeValue]; ok {
		return []int{conn[0], conn[1]}
	}
	conn := railConnections[r.RailShapeValue]
	return []int{conn[0], conn[1]}
}

func (r *Rail) GetPossibleConnectionDirectionsOneConstraint(constraint int) map[int]bool {
	possible := railDefaultPossibleConnectionDirectionsOneConstraint(constraint)
	if constraint&RailFlagAscend == 0 {
		for _, d := range math.HorizontalFacing {
			if constraint != int(d) {
				possible[int(d)] = true
			}
		}
	}
	return possible
}

func (r *Rail) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if railCanPlace(blockReplace) {
		return r.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}

func (r *Rail) OnPostPlace() { railOnPostPlace(r) }

func (r *Rail) OnNearbyBlockChange() { railOnNearbyBlockChange(r) }
