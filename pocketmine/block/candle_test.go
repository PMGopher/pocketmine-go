package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// candleWorld is fakeWorld with GetBlockAt returning a solid filler block, so
// GetAdjacentSupportType (used by Candle.Place) sees full support below.
type candleWorld struct {
	fakeWorld
}

func (w *candleWorld) GetBlockAt(x, y, z int) Behavior {
	filler := newTestBlock(false)
	filler.SetPosition(w, x, y, z)
	return filler
}

type fakeBlockTransaction struct{}

func (fakeBlockTransaction) AddBlock(pos Position, blk Behavior) {}

func newTestCandle(w World) *Candle {
	idInfo, err := NewBlockIdentifier(1004, nil)
	if err != nil {
		panic(err)
	}
	c := NewCandle(idInfo, "Test Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCandleFlintAndSteelLightsThenEmptyHandExtinguishes(t *testing.T) {
	w := &candleWorld{}
	c := newTestCandle(w)

	flintAndSteel := fakeItem{typeID: itemTypeIDsFlintAndSteel}
	if !c.OnInteract(flintAndSteel, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle flint and steel")
	}
	if !c.IsLit() {
		t.Error("flint and steel should have lit the candle")
	}

	emptyHand := fakeItem{null: true}
	if !c.OnInteract(emptyHand, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle the empty-hand extinguish")
	}
	if c.IsLit() {
		t.Error("empty hand should have extinguished the candle")
	}
}

func TestCandleStacksOnCompatibleCandle(t *testing.T) {
	w := &candleWorld{}
	existing := newTestCandle(w)
	existing.SetLit(true)

	newCandle := newTestCandle(w)
	tx := &fakeBlockTransaction{}
	if !newCandle.Place(tx, fakeItem{}, existing, existing, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed when stacking onto a compatible candle")
	}
	if newCandle.Count != 2 {
		t.Errorf("Count = %d, want 2 (should have incremented from the existing candle)", newCandle.Count)
	}
	if !newCandle.IsLit() {
		t.Error("expected the stacked candle to inherit the existing candle's lit state")
	}
}

func TestCandleCannotStackPastMax(t *testing.T) {
	w := &candleWorld{}
	existing := newTestCandle(w)
	existing.Count = candleMaxCount

	newCandle := newTestCandle(w)
	if newCandle.CanBePlacedAt(existing, math.Vector3{}, math.Up, true) {
		t.Error("expected CanBePlacedAt to reject stacking past the max count")
	}
}
