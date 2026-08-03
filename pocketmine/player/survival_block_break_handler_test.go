package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// fakeHandItem is a minimal block.Item test double - a bare hand (no tool type, efficiency 1.0,
// matching real Item::getMiningEfficiency's own default for an empty hand).
type fakeHandItem struct{}

func (fakeHandItem) GetTypeId() int                                        { return 0 }
func (fakeHandItem) GetBlockToolType() block.ToolType                      { return block.ToolTypeNone }
func (fakeHandItem) GetBlockToolHarvestLevel() int                         { return 0 }
func (fakeHandItem) GetMiningEfficiency(isCompatibleToolType bool) float64 { return 1 }
func (fakeHandItem) Pop()                                                  {}
func (fakeHandItem) IsNull() bool                                          { return false }
func (fakeHandItem) GetCustomName() string                                 { return "" }
func (fakeHandItem) GetCount() int                                         { return 1 }
func (fakeHandItem) SetCount(count int)                                    {}

func TestNewSurvivalBlockBreakHandlerComputesAPositiveBreakSpeedForABreakableBlock(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.ResetFallDistance()
	p.SetOnGround(true)

	h := NewSurvivalBlockBreakHandler(p, math.NewVector3(0, 70, 1), block.VanillaStone(), math.Up, 7, fakeHandItem{})
	if got := h.GetBreakSpeed(); got <= 0 {
		t.Errorf("GetBreakSpeed() = %v, want > 0 for a breakable block", got)
	}
}

func TestNewSurvivalBlockBreakHandlerIsZeroForAnUnbreakableBlock(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	h := NewSurvivalBlockBreakHandler(p, math.NewVector3(0, 70, 1), block.VanillaBedrock(), math.Up, 7, fakeHandItem{})
	if got := h.GetBreakSpeed(); got != 0 {
		t.Errorf("GetBreakSpeed() for bedrock (unbreakable) = %v, want 0", got)
	}
}

func TestUpdateAccumulatesBreakProgressUntilComplete(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetOnGround(true)

	h := NewSurvivalBlockBreakHandler(p, math.NewVector3(0, 70, 0), block.VanillaStone(), math.Up, 7, fakeHandItem{})

	ticks := 0
	for h.Update(fakeHandItem{}) {
		ticks++
		if ticks > 1000 {
			t.Fatal("break progress never reached 1 - looks like an infinite loop")
		}
	}
	if h.GetBreakProgress() < 1 {
		t.Errorf("GetBreakProgress() after Update() returned false = %v, want >= 1", h.GetBreakProgress())
	}
}

func TestUpdateReturnsFalseWhenPlayerIsTooFarAway(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	h := NewSurvivalBlockBreakHandler(p, math.NewVector3(100, 70, 100), block.VanillaStone(), math.Up, 7, fakeHandItem{})

	if h.Update(fakeHandItem{}) {
		t.Error("Update() = true despite the player being far outside maxPlayerDistance")
	}
}

func TestSetTargetedFacePanicsOnAnInvalidFacing(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	h := NewSurvivalBlockBreakHandler(p, math.NewVector3(0, 70, 1), block.VanillaStone(), math.Up, 7, fakeHandItem{})

	defer func() {
		if recover() == nil {
			t.Error("SetTargetedFace(999) did not panic")
		}
	}()
	h.SetTargetedFace(999)
}
