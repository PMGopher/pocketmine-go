package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"testing"
)

// noSupportBlock is a filler with SupportTypeNone, unlike testBlock/stemTestBlock which default
// to SupportTypeFull.
type noSupportBlock struct {
	Block
}

func (n *noSupportBlock) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (n *noSupportBlock) IsSolid() bool { return false }

func (n *noSupportBlock) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

// noSupportWorld returns a noSupportBlock filler for every GetBlockAt call.
type noSupportWorld struct {
	fakeWorld
}

func (w *noSupportWorld) GetBlockAt(x, y, z int) Behavior {
	n := &noSupportBlock{Block: NewBlock(mustBlockIdentifier(1034), "No Support Filler", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	n.Init(n)
	n.SetPosition(w, x, y, z)
	return n
}

func newTestCoral(w World) *Coral {
	c := NewCoral(mustBlockIdentifier(1033), "Test Coral", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCoralBreaksWithoutSupport(t *testing.T) {
	w := &noSupportWorld{}
	c := newTestCoral(w)

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}

func TestCoralSurvivesWithCenterSupportAndSchedulesUpdate(t *testing.T) {
	w := &candleWorld{}
	c := newTestCoral(w)

	c.OnNearbyBlockChange()

	if len(w.breakCalls) != 0 {
		t.Fatalf("expected no break with center support, got %d break calls", len(w.breakCalls))
	}
	if w.scheduleDelay < 40 || w.scheduleDelay > 200 {
		t.Errorf("scheduleDelay = %d, want in range [40, 200]", w.scheduleDelay)
	}
}

func TestCoralDeadDoesNotScheduleUpdate(t *testing.T) {
	w := &candleWorld{}
	c := newTestCoral(w)
	c.Dead = true

	c.OnNearbyBlockChange()

	if w.scheduleDelay != 0 {
		t.Errorf("expected no scheduled update while dead, got delay %d", w.scheduleDelay)
	}
}
