package block

import "testing"

// updateRecordingWorld is a fakeWorld that also supports World::setBlock's $update parameter.
type updateRecordingWorld struct {
	fakeWorld
	updates []bool
}

func (w *updateRecordingWorld) SetBlock(pos Position, blk Behavior) error {
	return w.SetBlockUpdate(pos, blk, true)
}

func (w *updateRecordingWorld) SetBlockUpdate(pos Position, blk Behavior, update bool) error {
	w.updates = append(w.updates, update)
	return w.fakeWorld.SetBlock(pos, blk)
}

func TestLeavesSetBlockWithoutNeighbourUpdates(t *testing.T) {
	w := &updateRecordingWorld{}
	leaves := VanillaOakLeaves().(*Leaves)
	leaves.position = NewPosition(0, 64, 0, w)

	// Leaves::onNearbyBlockChange: $world->setBlock($this->position, $this, false)
	leaves.OnNearbyBlockChange()
	if !leaves.CheckDecay {
		t.Fatal("OnNearbyBlockChange didn't set CheckDecay")
	}
	if len(w.updates) != 1 || w.updates[0] {
		t.Errorf("setBlock calls with update = %v, want [false]", w.updates)
	}
}
