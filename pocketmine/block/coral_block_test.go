package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestCoralBlock(w World) *CoralBlock {
	c := NewCoralBlock(mustBlockIdentifier(1035), "Test Coral Block", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCoralBlockOnScheduledUpdateDiesWithoutWater(t *testing.T) {
	w := &candleWorld{}
	c := newTestCoralBlock(w)

	c.OnScheduledUpdate()

	dead, ok := w.lastSetBlock.(*CoralBlock)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *CoralBlock, got %T", w.lastSetBlock)
	}
	if !dead.Dead {
		t.Error("expected the replacement coral block to be dead")
	}
}

func TestCoralBlockOnScheduledUpdateSurvivesAdjacentToWater(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCoralBlock(w)
	origin := math.Vector3{X: 1, Y: 2, Z: 3}
	side := origin.GetSide(math.East, 1)
	water := &stemTestBlock{typeID: WATER}
	water.Block = NewBlock(mustBlockIdentifier(1041), "Water Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	water.Init(water)
	water.SetPosition(w, side.FloorX(), side.FloorY(), side.FloorZ())
	w.blocks[[3]int{side.FloorX(), side.FloorY(), side.FloorZ()}] = water

	c.OnScheduledUpdate()

	if w.lastSetBlock != nil {
		t.Errorf("expected no death when adjacent to water, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestCoralBlockOnScheduledUpdateDoesNothingWhenAlreadyDead(t *testing.T) {
	w := &candleWorld{}
	c := newTestCoralBlock(w)
	c.Dead = true

	c.OnScheduledUpdate()

	if w.lastSetBlock != nil {
		t.Errorf("expected no SetBlock while already dead, got SetBlock(%T)", w.lastSetBlock)
	}
}
