package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

// testCropsLeaf is a minimal concrete Crops leaf for testing - Crops itself is never instantiated
// directly (no real leaf type is ported yet; Wheat/Carrot/Potato/Beetroot all need the item
// registry for their AsItem()/GetDropsForCompatibleTool overrides), same convention as testBlock
// in behavior_test.go.
type testCropsLeaf struct {
	Crops
}

func newTestCropsLeaf(w World) *testCropsLeaf {
	c := &testCropsLeaf{Crops: Crops{
		Flowable:     Flowable{Transparent{NewBlock(mustBlockIdentifier(1096), "Test Crops", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}},
		AgeComponent: NewAgeComponent(CropsMaxAge),
	}}
	c.Init(c)
	c.SetPosition(w, 1, 2, 3)
	return c
}

func (c *testCropsLeaf) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func TestCropsOnInteractFertilizesWithBoneMeal(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCropsLeaf(w)
	c.Age = 0
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	if !c.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	grown, ok := w.lastSetBlock.(*testCropsLeaf)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *testCropsLeaf, got %T", w.lastSetBlock)
	}
	if grown.Age < 2 || grown.Age > 5 {
		t.Errorf("Age = %d, want in range 2-5", grown.Age)
	}
}

func TestCropsOnInteractFertilizeClampsToMaxAge(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCropsLeaf(w)
	c.Age = CropsMaxAge - 1
	boneMeal := fakeItem{typeID: itemTypeIDsBoneMeal}

	c.OnInteract(boneMeal, math.Up, math.Vector3{}, nil, nil)

	grown, ok := w.lastSetBlock.(*testCropsLeaf)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *testCropsLeaf, got %T", w.lastSetBlock)
	}
	if grown.Age != CropsMaxAge {
		t.Errorf("Age = %d, want clamped to CropsMaxAge (%d)", grown.Age, CropsMaxAge)
	}
}

func TestCropsOnInteractIgnoresNonFertilizerItems(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCropsLeaf(w)

	if c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to return false for a non-fertilizer item")
	}
	if w.lastSetBlock != nil {
		t.Error("expected no state change for a non-fertilizer item")
	}
}

func TestCropsTicksRandomlyOnlyBelowMaxAge(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCropsLeaf(w)
	c.Age = CropsMaxAge - 1
	if !c.TicksRandomly() {
		t.Error("expected TicksRandomly() to be true below max age")
	}
	c.Age = CropsMaxAge
	if c.TicksRandomly() {
		t.Error("expected TicksRandomly() to be false at max age")
	}
}

func TestCropsOnRandomTickDoesNothingAtMaxAge(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCropsLeaf(w)
	c.Age = CropsMaxAge

	c.OnRandomTick()

	if w.lastSetBlock != nil {
		t.Error("expected no growth at max age")
	}
}
