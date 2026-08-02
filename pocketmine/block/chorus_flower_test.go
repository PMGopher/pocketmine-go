package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// chorusFlowerWorld is a coordinate-keyed World double with a non-air, non-replaceable default
// filler (so "clear" spots must be placed explicitly), IsInWorld overridable per coordinate, and
// every SetBlock call recorded (reusing cactusSetCall/AddSound/sounds already on fakeWorld).
type chorusFlowerWorld struct {
	fakeWorld
	blocks     map[[3]int]Behavior
	outOfWorld map[[3]int]bool
	setCalls   []cactusSetCall
}

func newChorusFlowerWorld() *chorusFlowerWorld {
	return &chorusFlowerWorld{blocks: map[[3]int]Behavior{}, outOfWorld: map[[3]int]bool{}}
}

func (w *chorusFlowerWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newFakeTypedBlock(STONE)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *chorusFlowerWorld) IsInWorld(x, y, z int) bool { return !w.outOfWorld[[3]int{x, y, z}] }

func (w *chorusFlowerWorld) SetBlock(pos Position, blk Behavior) error {
	w.setCalls = append(w.setCalls, cactusSetCall{pos, blk})
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
}

func placeTypedAt(w *chorusFlowerWorld, x, y, z, typeID int) {
	b := newFakeTypedBlockAt(w, typeID, x, y, z)
	w.blocks[[3]int{x, y, z}] = b
}

func newTestChorusFlower(w World) *ChorusFlower {
	c := NewChorusFlower(mustBlockIdentifier(CHORUS_FLOWER), "Test Chorus Flower", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 5, 5, 5)
	return c
}

func TestChorusFlowerScanStemCountsConsecutiveChorusPlant(t *testing.T) {
	w := newChorusFlowerWorld()
	placeTypedAt(w, 5, 4, 5, CHORUS_PLANT)
	placeTypedAt(w, 5, 3, 5, CHORUS_PLANT)
	placeTypedAt(w, 5, 2, 5, STONE)
	c := newTestChorusFlower(w)

	height, endStone := c.scanStem()
	if height != 2 {
		t.Errorf("stemHeight = %d, want 2", height)
	}
	if endStone {
		t.Error("expected endStoneBelow = false")
	}
}

func TestChorusFlowerScanStemDetectsEndStoneBelow(t *testing.T) {
	w := newChorusFlowerWorld()
	placeTypedAt(w, 5, 4, 5, END_STONE)
	c := newTestChorusFlower(w)

	height, endStone := c.scanStem()
	if height != 0 {
		t.Errorf("stemHeight = %d, want 0", height)
	}
	if !endStone {
		t.Error("expected endStoneBelow = true")
	}
}

func TestChorusFlowerAllHorizontalBlocksEmptyTrueWhenAllAir(t *testing.T) {
	w := newChorusFlowerWorld()
	for _, face := range math.HorizontalFacing {
		side := math.Vector3{X: 5, Y: 5, Z: 5}.GetSide(face, 1)
		placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), AIR)
	}
	c := newTestChorusFlower(w)

	if !c.allHorizontalBlocksEmpty(w, math.Vector3{X: 5, Y: 5, Z: 5}, nil) {
		t.Error("expected all horizontal neighbours to read as empty")
	}
}

func TestChorusFlowerAllHorizontalBlocksEmptyFalseWhenOneBlocked(t *testing.T) {
	w := newChorusFlowerWorld()
	for _, face := range math.HorizontalFacing {
		side := math.Vector3{X: 5, Y: 5, Z: 5}.GetSide(face, 1)
		placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), AIR)
	}
	// Block the North side.
	side := math.Vector3{X: 5, Y: 5, Z: 5}.GetSide(math.North, 1)
	placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), STONE)
	c := newTestChorusFlower(w)

	if c.allHorizontalBlocksEmpty(w, math.Vector3{X: 5, Y: 5, Z: 5}, nil) {
		t.Error("expected a blocked North side to fail the check")
	}
}

func TestChorusFlowerAllHorizontalBlocksEmptyIgnoresExceptedFacing(t *testing.T) {
	w := newChorusFlowerWorld()
	for _, face := range math.HorizontalFacing {
		side := math.Vector3{X: 5, Y: 5, Z: 5}.GetSide(face, 1)
		placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), AIR)
	}
	side := math.Vector3{X: 5, Y: 5, Z: 5}.GetSide(math.North, 1)
	placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), STONE)
	c := newTestChorusFlower(w)

	except := math.North
	if !c.allHorizontalBlocksEmpty(w, math.Vector3{X: 5, Y: 5, Z: 5}, &except) {
		t.Error("expected the excepted (North) side to be skipped, leaving the check passing")
	}
}

// setupClearAbove places AIR at Y+1 and Y+2, AIR directly below the flower (so canGrowUpwards
// never enters the randomness-gated stemHeight branch), and AIR on every horizontal neighbour of
// Y+1 - a fully deterministic "can grow upwards" environment.
func setupClearAbove(w *chorusFlowerWorld) {
	placeTypedAt(w, 5, 6, 5, AIR) // up
	placeTypedAt(w, 5, 7, 5, AIR) // up.up
	placeTypedAt(w, 5, 4, 5, AIR) // below the flower itself
	for _, face := range math.HorizontalFacing {
		side := math.Vector3{X: 5, Y: 6, Z: 5}.GetSide(face, 1)
		placeTypedAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), AIR)
	}
}

func TestChorusFlowerCanGrowUpwardsTrueWhenClear(t *testing.T) {
	w := newChorusFlowerWorld()
	setupClearAbove(w)
	c := newTestChorusFlower(w)

	if !c.canGrowUpwards(0, false) {
		t.Error("expected canGrowUpwards to succeed in a fully clear environment")
	}
}

func TestChorusFlowerCanGrowUpwardsFalseWhenSpaceAboveBlocked(t *testing.T) {
	w := newChorusFlowerWorld()
	setupClearAbove(w)
	placeTypedAt(w, 5, 6, 5, STONE) // block "up"
	c := newTestChorusFlower(w)

	if c.canGrowUpwards(0, false) {
		t.Error("expected canGrowUpwards to fail when the space directly above is blocked")
	}
}

func TestChorusFlowerCanGrowUpwardsFalseAtMaxStemHeight(t *testing.T) {
	w := newChorusFlowerWorld()
	setupClearAbove(w)
	placeTypedAt(w, 5, 4, 5, STONE) // below not air, so the stem-height check is reached
	c := newTestChorusFlower(w)

	if c.canGrowUpwards(chorusFlowerMaxStemHeight, false) {
		t.Error("expected canGrowUpwards to fail once stemHeight reaches the max")
	}
}

func TestChorusFlowerGrowAddsClonedBlockWithClampedAge(t *testing.T) {
	w := newChorusFlowerWorld()
	c := newTestChorusFlower(w)
	c.Age = ChorusFlowerMaxAge - 1

	tx := c.grow(math.Up, 5, nil)
	if tx == nil {
		t.Fatal("expected grow to return a non-nil transaction")
	}
	if !tx.Apply() {
		t.Fatal("expected Apply to report a change")
	}
	grown, ok := w.setCalls[0].blk.(*ChorusFlower)
	if !ok {
		t.Fatalf("expected a *ChorusFlower, got %T", w.setCalls[0].blk)
	}
	if grown.Age != ChorusFlowerMaxAge {
		t.Errorf("Age = %d, want clamped to %d", grown.Age, ChorusFlowerMaxAge)
	}
	if w.setCalls[0].pos.FloorY() != 6 {
		t.Errorf("Y = %d, want 6 (one step up)", w.setCalls[0].pos.FloorY())
	}
}

func TestChorusFlowerOnRandomTickDoesNothingAtMaxAge(t *testing.T) {
	w := newChorusFlowerWorld()
	c := newTestChorusFlower(w)
	c.Age = ChorusFlowerMaxAge

	c.OnRandomTick()

	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestChorusFlowerOnRandomTickDiesWhenNothingCanGrow(t *testing.T) {
	// Default filler (STONE, opaque) everywhere blocks both the upward path (up isn't AIR) and
	// every horizontal attempt (sides aren't AIR either), regardless of the random roll/attempt
	// count/facing choices.
	w := newChorusFlowerWorld()
	c := newTestChorusFlower(w)

	c.OnRandomTick()

	if len(w.setCalls) != 1 {
		t.Fatalf("expected exactly 1 SetBlock call (the flower persisting at max age), got %d", len(w.setCalls))
	}
	dead, ok := w.setCalls[0].blk.(*ChorusFlower)
	if !ok || dead.Age != ChorusFlowerMaxAge {
		t.Errorf("expected a *ChorusFlower at max age, got %#v", w.setCalls[0].blk)
	}
	if len(w.sounds) != 1 {
		t.Errorf("expected 1 sound to play, got %d", len(w.sounds))
	}
}

func TestChorusFlowerOnRandomTickGrowsUpwardWhenClear(t *testing.T) {
	w := newChorusFlowerWorld()
	setupClearAbove(w)
	c := newTestChorusFlower(w)

	c.OnRandomTick()

	if len(w.setCalls) != 2 {
		t.Fatalf("expected exactly 2 SetBlock calls (new flower above + chorus plant at self), got %d", len(w.setCalls))
	}
	_, ok := w.setCalls[0].blk.(*ChorusFlower)
	if !ok || w.setCalls[0].pos.FloorY() != 6 {
		t.Errorf("expected a *ChorusFlower at Y=6, got %#v at Y=%d", w.setCalls[0].blk, w.setCalls[0].pos.FloorY())
	}
	_, ok = w.setCalls[1].blk.(*ChorusPlant)
	if !ok || w.setCalls[1].pos.FloorY() != 5 {
		t.Errorf("expected a *ChorusPlant at Y=5 (the flower's own position), got %#v at Y=%d", w.setCalls[1].blk, w.setCalls[1].pos.FloorY())
	}
}
