package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestRedstoneRepeaterOnInteractCyclesDelayAndWrapsToMin(t *testing.T) {
	w := &fakeWorld{}
	r := NewRedstoneRepeater(mustBlockIdentifier(1026), "Test Repeater", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	r.SetPosition(w, 1, 2, 3)

	for expected := 2; expected <= redstoneRepeaterMaxDelay; expected++ {
		if !r.OnInteract(fakeItem{}, 0, math.Vector3{}, nil, nil) {
			t.Fatal("expected OnInteract to report handled")
		}
		if r.Delay != expected {
			t.Errorf("Delay = %d, want %d", r.Delay, expected)
		}
	}

	// One more interaction past max should wrap back to min.
	r.OnInteract(fakeItem{}, 0, math.Vector3{}, nil, nil)
	if r.Delay != redstoneRepeaterMinDelay {
		t.Errorf("Delay = %d, want %d (wrap to min)", r.Delay, redstoneRepeaterMinDelay)
	}
}

func TestRedstoneComparatorOnInteractTogglesSubtractMode(t *testing.T) {
	w := &fakeWorld{}
	c := NewRedstoneComparator(mustBlockIdentifier(1027), "Test Comparator", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)

	if c.IsSubtractMode() {
		t.Fatal("expected default SubtractMode to be false")
	}
	c.OnInteract(fakeItem{}, 0, math.Vector3{}, nil, nil)
	if !c.IsSubtractMode() {
		t.Error("expected SubtractMode to be true after one interaction")
	}
	c.OnInteract(fakeItem{}, 0, math.Vector3{}, nil, nil)
	if c.IsSubtractMode() {
		t.Error("expected SubtractMode to be false after a second interaction")
	}
}
