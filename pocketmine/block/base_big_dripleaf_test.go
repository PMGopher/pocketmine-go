package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestBaseBigDripleafPlaceConvertsDripleafBelowIntoStem(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	below := newTestBigDripleafHead(w)
	below.SetPosition(w, 1, 1, 3)
	below.SetFacing(math.East)
	w.blocks[[3]int{1, 1, 3}] = below

	replace := newTestBlock(true)
	replace.SetPosition(w, 1, 2, 3)

	stem := NewBigDripleafStem(mustBlockIdentifier(1100), "Test Stem To Place", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	stem.SetPosition(w, 1, 2, 3)

	tx := &recordingTransaction{}
	if !stem.Place(tx, fakeItem{}, replace, replace, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed")
	}

	if len(tx.blocks) != 2 {
		t.Fatalf("expected 2 blocks added to the transaction (converted stem below + the placed stem itself), got %d", len(tx.blocks))
	}
	converted, ok := tx.blocks[0].(*BigDripleafStem)
	if !ok {
		t.Fatalf("expected the first added block to be a *BigDripleafStem, got %T", tx.blocks[0])
	}
	if converted.GetFacing() != math.East {
		t.Errorf("converted stem Facing = %v, want East (carried over from the head below)", converted.GetFacing())
	}
	if stem.Facing != math.East {
		t.Errorf("placed stem's own Facing = %v, want East (carried from the head below)", stem.Facing)
	}
}

func TestBaseBigDripleafPlaceDoesNotConvertNonDripleafBelow(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	clayBelow := newFakeTypedBlockAt(w, CLAY, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = clayBelow

	replace := newTestBlock(true)
	replace.SetPosition(w, 1, 2, 3)

	stem := NewBigDripleafStem(mustBlockIdentifier(1100), "Test Stem To Place", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	stem.SetPosition(w, 1, 2, 3)

	tx := &recordingTransaction{}
	if !stem.Place(tx, fakeItem{}, replace, replace, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed against clay support")
	}

	if len(tx.blocks) != 1 {
		t.Fatalf("expected only the placed stem itself in the transaction, got %d blocks", len(tx.blocks))
	}
	if _, ok := tx.blocks[0].(*BigDripleafStem); !ok {
		t.Fatalf("expected the sole added block to be the placed *BigDripleafStem, got %T", tx.blocks[0])
	}
}
