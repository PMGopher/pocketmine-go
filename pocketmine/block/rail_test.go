package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// fakeRailWorld is a minimal World backed by a position->block map, used to exercise the rail
// auto-connect algorithm (railTryReconnect) against real neighbor lookups. Positions with no
// entry return a plain non-rail testBlock, standing in for "some other block" / effectively air
// for connection purposes.
type fakeRailWorld struct {
	blocks map[[3]int]Behavior
}

func (w *fakeRailWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(false)
	filler.SetPosition(w, x, y, z)
	return filler
}
func (w *fakeRailWorld) SetBlock(pos Position, blk Behavior) error {
	w.blocks[[3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}] = blk
	return nil
}
func (w *fakeRailWorld) GetTile(pos Position) (Tile, bool)                      { return nil, false }
func (w *fakeRailWorld) AddTile(tile Tile)                                      {}
func (w *fakeRailWorld) GetOrLoadChunkAtPosition(pos Position) (Chunk, bool)    { return nil, false }
func (w *fakeRailWorld) AddSound(pos math.Vector3, s sound.Sound)               {}
func (w *fakeRailWorld) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {}
func (w *fakeRailWorld) GetFullLightAt(x, y, z int) int                         { return 15 }
func (w *fakeRailWorld) GetBlockLightAt(x, y, z int) int                        { return 15 }
func (w *fakeRailWorld) GetRealBlockSkyLightAt(x, y, z int) int                 { return 15 }
func (w *fakeRailWorld) GetSunAnglePercentage() float64                         { return 0.5 }
func (w *fakeRailWorld) GetNearbyEntities(bb math.AxisAlignedBB) []Entity       { return nil }
func (w *fakeRailWorld) UseBreakOn(pos math.Vector3) bool                       { return true }

func newTestRailAt(w *fakeRailWorld, x, y, z int) *Rail {
	idInfo, err := NewBlockIdentifier(2000, nil)
	if err != nil {
		panic(err)
	}
	r := NewRail(idInfo, "Test Rail", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	r.SetPosition(w, x, y, z)
	w.blocks[[3]int{x, y, z}] = r
	return r
}

func TestRailAutoConnectsToAdjacentRail(t *testing.T) {
	w := &fakeRailWorld{blocks: map[[3]int]Behavior{}}

	railA := newTestRailAt(w, 0, 0, 0)
	railB := newTestRailAt(w, 1, 0, 0) // east of A

	railB.OnPostPlace()

	if railA.GetShape() != RailStraightEastWest {
		t.Errorf("railA.GetShape() = %d, want RailStraightEastWest (%d)", railA.GetShape(), RailStraightEastWest)
	}
	if railB.GetShape() != RailStraightEastWest {
		t.Errorf("railB.GetShape() = %d, want RailStraightEastWest (%d)", railB.GetShape(), RailStraightEastWest)
	}
}

func TestRailSearchStateFindsShapeInEitherOrder(t *testing.T) {
	shape, ok := railSearchState([]int{int(math.West), int(math.East)}, railConnections)
	if !ok || shape != RailStraightEastWest {
		t.Errorf("got (%d, %v), want (%d, true)", shape, ok, RailStraightEastWest)
	}

	// Reversed order should still match.
	shape, ok = railSearchState([]int{int(math.East), int(math.West)}, railConnections)
	if !ok || shape != RailStraightEastWest {
		t.Errorf("reversed: got (%d, %v), want (%d, true)", shape, ok, RailStraightEastWest)
	}

	shape, ok = railSearchState([]int{int(math.South), int(math.East)}, railCurveConnections)
	if !ok || shape != RailCurveSoutheast {
		t.Errorf("curve: got (%d, %v), want (%d, true)", shape, ok, RailCurveSoutheast)
	}
}
