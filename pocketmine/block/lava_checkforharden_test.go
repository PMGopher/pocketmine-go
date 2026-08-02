package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// lavaHardenWorld returns a fixed type ID for every GetBlockAt call except at positions
// explicitly overridden via blocks - used to exercise Lava.checkForHarden's neighbour checks
// without the block registry (same pattern as stemWorld/sugarcaneWorld).
type lavaHardenWorld struct {
	fakeWorld
	fillerTypeID int
	blocks       map[[3]int]Behavior
}

func (w *lavaHardenWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := &stemTestBlock{typeID: w.fillerTypeID}
	filler.Block = NewBlock(mustBlockIdentifier(1043), "Lava Harden Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	filler.Init(filler)
	filler.SetPosition(w, x, y, z)
	return filler
}

func newLavaHardenTestLava(w World) *Lava {
	l := NewLava(mustBlockIdentifier(1042), "Test Lava", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	l.SetPosition(w, 1, 2, 3)
	return l
}

func placeAt(w *lavaHardenWorld, x, y, z int, typeID int) {
	b := &stemTestBlock{typeID: typeID}
	b.Block = NewBlock(mustBlockIdentifier(1044), "Placed", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.Init(b)
	b.SetPosition(w, x, y, z)
	w.blocks[[3]int{x, y, z}] = b
}

func TestLavaCheckForHardenReturnsFalseWhileFalling(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)
	l.Falling = true
	side := math.Vector3{X: 1, Y: 2, Z: 3}.GetSide(math.North, 1)
	waterBlk := newTestWater(w)
	waterBlk.SetPosition(w, side.FloorX(), side.FloorY(), side.FloorZ())
	w.blocks[[3]int{side.FloorX(), side.FloorY(), side.FloorZ()}] = waterBlk

	if l.checkForHarden() {
		t.Error("expected checkForHarden to return false while falling")
	}
	if w.lastSetBlock != nil {
		t.Errorf("expected no SetBlock while falling, got %T", w.lastSetBlock)
	}
}

func TestLavaCheckForHardenSourceWaterTurnsToObsidian(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)
	l.Decay = 0
	side := math.Vector3{X: 1, Y: 2, Z: 3}.GetSide(math.North, 1)
	waterBlk := newTestWater(w)
	waterBlk.SetPosition(w, side.FloorX(), side.FloorY(), side.FloorZ())
	w.blocks[[3]int{side.FloorX(), side.FloorY(), side.FloorZ()}] = waterBlk

	if !l.checkForHarden() {
		t.Fatal("expected checkForHarden to return true")
	}
	obsidian, ok := w.lastSetBlock.(*Opaque)
	if !ok || obsidian.GetTypeId() != OBSIDIAN {
		t.Fatalf("expected SetBlock to be called with Obsidian, got %#v", w.lastSetBlock)
	}
}

func TestLavaCheckForHardenFlowingWaterLowDecayTurnsToCobblestone(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)
	l.Decay = 3
	side := math.Vector3{X: 1, Y: 2, Z: 3}.GetSide(math.East, 1)
	waterBlk := newTestWater(w)
	waterBlk.SetPosition(w, side.FloorX(), side.FloorY(), side.FloorZ())
	w.blocks[[3]int{side.FloorX(), side.FloorY(), side.FloorZ()}] = waterBlk

	if !l.checkForHarden() {
		t.Fatal("expected checkForHarden to return true")
	}
	cobble, ok := w.lastSetBlock.(*Opaque)
	if !ok || cobble.GetTypeId() != COBBLESTONE {
		t.Fatalf("expected SetBlock to be called with Cobblestone, got %#v", w.lastSetBlock)
	}
}

func TestLavaCheckForHardenHighDecayWaterDoesNotHarden(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)
	l.Decay = 5
	side := math.Vector3{X: 1, Y: 2, Z: 3}.GetSide(math.East, 1)
	waterBlk := newTestWater(w)
	waterBlk.SetPosition(w, side.FloorX(), side.FloorY(), side.FloorZ())
	w.blocks[[3]int{side.FloorX(), side.FloorY(), side.FloorZ()}] = waterBlk

	if l.checkForHarden() {
		t.Error("expected checkForHarden to return false at decay 5")
	}
	if w.lastSetBlock != nil {
		t.Errorf("expected no SetBlock at decay 5, got %T", w.lastSetBlock)
	}
}

func TestLavaCheckForHardenBlueIceOnSoulSoilTurnsToBasalt(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)
	origin := math.Vector3{X: 1, Y: 2, Z: 3}
	down := origin.GetSide(math.Down, 1)
	placeAt(w, down.FloorX(), down.FloorY(), down.FloorZ(), SOUL_SOIL)
	side := origin.GetSide(math.East, 1)
	placeAt(w, side.FloorX(), side.FloorY(), side.FloorZ(), BLUE_ICE)

	if !l.checkForHarden() {
		t.Fatal("expected checkForHarden to return true")
	}
	basalt, ok := w.lastSetBlock.(*SimplePillar)
	if !ok || basalt.GetTypeId() != BASALT {
		t.Fatalf("expected SetBlock to be called with Basalt, got %#v", w.lastSetBlock)
	}
}

func TestLavaCheckForHardenNoRelevantNeighboursDoesNothing(t *testing.T) {
	w := &lavaHardenWorld{fillerTypeID: STONE, blocks: map[[3]int]Behavior{}}
	l := newLavaHardenTestLava(w)

	if l.checkForHarden() {
		t.Error("expected checkForHarden to return false without water or blue ice")
	}
	if w.lastSetBlock != nil {
		t.Errorf("expected no SetBlock, got %T", w.lastSetBlock)
	}
}
