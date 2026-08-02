package block

import "testing"

// btWorld is a minimal World double for BlockTransactionImpl tests: a coordinate-keyed block map,
// a per-coordinate "out of world" override, and a record of every SetBlock call.
type btWorld struct {
	fakeWorld
	blocks     map[[3]int]Behavior
	outOfWorld map[[3]int]bool
	setCalls   []cactusSetCall
}

func (w *btWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(true)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *btWorld) IsInWorld(x, y, z int) bool { return !w.outOfWorld[[3]int{x, y, z}] }

func (w *btWorld) SetBlock(pos Position, blk Behavior) error {
	w.setCalls = append(w.setCalls, cactusSetCall{pos, blk})
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
}

func newBTWorld() *btWorld {
	return &btWorld{blocks: map[[3]int]Behavior{}, outOfWorld: map[[3]int]bool{}}
}

func TestBlockTransactionApplyAppliesAllBlocksAndReturnsTrue(t *testing.T) {
	w := newBTWorld()
	tx := NewBlockTransaction(w)
	a := newFakeTypedBlockAt(w, STONE, 1, 2, 3)
	b := newFakeTypedBlockAt(w, DIRT, 4, 5, 6)
	tx.AddBlockAt(1, 2, 3, a)
	tx.AddBlockAt(4, 5, 6, b)

	if !tx.Apply() {
		t.Fatal("expected Apply to return true")
	}
	if len(w.setCalls) != 2 {
		t.Fatalf("expected 2 SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestBlockTransactionApplySkipsUnchangedBlocksAndReturnsFalse(t *testing.T) {
	w := newBTWorld()
	existing := newFakeTypedBlockAt(w, STONE, 1, 2, 3)
	w.blocks[[3]int{1, 2, 3}] = existing

	tx := NewBlockTransaction(w)
	// Same type ID (and thus same state ID) as what's already there.
	tx.AddBlockAt(1, 2, 3, newFakeTypedBlockAt(w, STONE, 1, 2, 3))

	if tx.Apply() {
		t.Error("expected Apply to return false when nothing actually changed")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls for an unchanged block, got %d", len(w.setCalls))
	}
}

func TestBlockTransactionApplyFailsAtomicallyWhenAnyPositionOutOfWorld(t *testing.T) {
	w := newBTWorld()
	w.outOfWorld[[3]int{4, 5, 6}] = true

	tx := NewBlockTransaction(w)
	tx.AddBlockAt(1, 2, 3, newFakeTypedBlockAt(w, STONE, 1, 2, 3))
	tx.AddBlockAt(4, 5, 6, newFakeTypedBlockAt(w, DIRT, 4, 5, 6))

	if tx.Apply() {
		t.Error("expected Apply to return false when any position is out of world")
	}
	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls at all when validation fails, got %d", len(w.setCalls))
	}
}

func TestBlockTransactionFetchBlockAtReturnsPendingOverWorld(t *testing.T) {
	w := newBTWorld()
	w.blocks[[3]int{1, 2, 3}] = newFakeTypedBlockAt(w, STONE, 1, 2, 3)

	tx := NewBlockTransaction(w)
	if got := tx.FetchBlockAt(1, 2, 3); got.GetTypeId() != STONE {
		t.Errorf("expected FetchBlockAt to fall back to the world block, got type %d", got.GetTypeId())
	}

	pending := newFakeTypedBlockAt(w, DIRT, 1, 2, 3)
	tx.AddBlockAt(1, 2, 3, pending)
	if got := tx.FetchBlockAt(1, 2, 3); got.GetTypeId() != DIRT {
		t.Errorf("expected FetchBlockAt to prefer the pending block, got type %d", got.GetTypeId())
	}
}

func TestBlockTransactionAddBlockAtOverwritesSamePosition(t *testing.T) {
	w := newBTWorld()
	tx := NewBlockTransaction(w)
	tx.AddBlockAt(1, 2, 3, newFakeTypedBlockAt(w, STONE, 1, 2, 3))
	tx.AddBlockAt(1, 2, 3, newFakeTypedBlockAt(w, DIRT, 1, 2, 3))

	if !tx.Apply() {
		t.Fatal("expected Apply to return true")
	}
	if len(w.setCalls) != 1 {
		t.Fatalf("expected exactly 1 SetBlock call (second AddBlockAt overwrites the first), got %d", len(w.setCalls))
	}
	if w.setCalls[0].blk.GetTypeId() != DIRT {
		t.Errorf("expected the final overwrite (DIRT) to win, got type %d", w.setCalls[0].blk.GetTypeId())
	}
}
