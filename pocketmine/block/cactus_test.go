package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func newTestCactus(w World) *Cactus {
	c := NewCactus(mustBlockIdentifier(1086), "Test Cactus", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCactusOnEntityInsideDealsContactDamage(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCactus(w)

	e := entity.NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := e.GetHealth()

	if !c.OnEntityInside(e) {
		t.Fatal("expected OnEntityInside to return true")
	}
	if e.GetHealth() != startHealth-1 {
		t.Errorf("GetHealth() = %v, want %v", e.GetHealth(), startHealth-1)
	}
	if e.GetLastDamageCause().GetCause() != entity.EntityDamageCauseContact {
		t.Errorf("GetCause() = %d, want EntityDamageCauseContact", e.GetLastDamageCause().GetCause())
	}
}

type cactusSetCall struct {
	pos Position
	blk Behavior
}

// cactusTickWorld records every SetBlock call (in order) so Cactus.OnRandomTick's multiple
// possible SetBlock calls (the Grow() replacement at "up", then the self-persist at its own
// position) can both be asserted, unlike fakeWorld which only remembers the last one.
type cactusTickWorld struct {
	fakeWorld
	blocks   map[[3]int]Behavior
	inWorld  bool
	setCalls []cactusSetCall
}

func (w *cactusTickWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(true)
	filler.SetPosition(w, x, y, z)
	return filler
}

func (w *cactusTickWorld) IsInWorld(x, y, z int) bool { return w.inWorld }

func (w *cactusTickWorld) SetBlock(pos Position, blk Behavior) error {
	w.setCalls = append(w.setCalls, cactusSetCall{pos, blk})
	w.lastSetPos, w.lastSetBlock = pos, blk
	return nil
}

func newFakeTypedBlockAt(w World, typeID, x, y, z int) *fakeTypedBlock {
	b := newFakeTypedBlock(typeID)
	b.SetPosition(w, x, y, z)
	return b
}

func TestCactusOnRandomTickDoesNothingWhenAboveIsNotAir(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: true}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, STONE, 1, 3, 3)
	c := newTestCactus(w)

	c.OnRandomTick()

	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestCactusOnRandomTickDoesNothingWhenAboveIsOutOfWorld(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: false}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, AIR, 1, 3, 3)
	c := newTestCactus(w)

	c.OnRandomTick()

	if len(w.setCalls) != 0 {
		t.Errorf("expected no SetBlock calls, got %d", len(w.setCalls))
	}
}

func TestCactusOnRandomTickIncrementsAgeBelowMax(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: true}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, AIR, 1, 3, 3)
	c := newTestCactus(w)
	c.Age = 5

	c.OnRandomTick()

	if len(w.setCalls) != 1 {
		t.Fatalf("expected exactly 1 SetBlock call, got %d", len(w.setCalls))
	}
	if c.Age != 6 {
		t.Errorf("Age = %d, want 6", c.Age)
	}
	if w.setCalls[0].blk != Behavior(c) {
		t.Error("expected the self-persist SetBlock to store the cactus itself")
	}
}

func TestCactusOnRandomTickGrowsUpwardAtMaxAgeBelowMaxHeight(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: true}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, AIR, 1, 3, 3)
	c := newTestCactus(w)
	c.Age = CactusMaxAge

	c.OnRandomTick()

	if len(w.setCalls) != 2 {
		t.Fatalf("expected exactly 2 SetBlock calls (grow + self-persist), got %d", len(w.setCalls))
	}
	grown, ok := w.setCalls[0].blk.(*Cactus)
	if !ok {
		t.Fatalf("expected the first SetBlock to place a *Cactus above, got %T", w.setCalls[0].blk)
	}
	if grown == c {
		t.Error("expected the grown replacement to be a distinct instance, not the original")
	}
	if w.setCalls[0].pos.FloorY() != 3 {
		t.Errorf("grow position Y = %d, want 3 (above the cactus)", w.setCalls[0].pos.FloorY())
	}
	if c.Age != 0 {
		t.Errorf("Age = %d, want reset to 0", c.Age)
	}
}

func TestCactusOnRandomTickNoUpwardGrowthAtMaxHeight(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: true}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, AIR, 1, 3, 3)
	// Stack two more cacti below (1,1,3) and (1,0,3) so height reaches CactusMaxHeight (3).
	// Uses type ID 1086 to match newTestCactus's own (non-canonical) block identifier, since
	// HasSameTypeId compares against that, not the real CACTUS constant.
	w.blocks[[3]int{1, 1, 3}] = newFakeTypedBlockAt(w, 1086, 1, 1, 3)
	w.blocks[[3]int{1, 0, 3}] = newFakeTypedBlockAt(w, 1086, 1, 0, 3)
	c := newTestCactus(w)
	c.Age = CactusMaxAge

	c.OnRandomTick()

	if len(w.setCalls) != 1 {
		t.Fatalf("expected exactly 1 SetBlock call (self-persist only), got %d", len(w.setCalls))
	}
	if c.Age != 0 {
		t.Errorf("Age = %d, want reset to 0", c.Age)
	}
}

func TestCactusOnRandomTickAtNineDoesNotAttemptFlowerWhenSurrounded(t *testing.T) {
	w := &cactusTickWorld{blocks: map[[3]int]Behavior{}, inWorld: true}
	w.blocks[[3]int{1, 3, 3}] = newFakeTypedBlockAt(w, AIR, 1, 3, 3)
	// Surround the block above with solid neighbours so canGrowFlower is false.
	solid := newTestBlock(false)
	solid.SetPosition(w, 2, 3, 3)
	w.blocks[[3]int{2, 3, 3}] = solid
	c := newTestCactus(w)
	c.Age = 9

	c.OnRandomTick()

	if len(w.setCalls) != 1 {
		t.Fatalf("expected exactly 1 SetBlock call (age increment, no flower attempt), got %d", len(w.setCalls))
	}
	if c.Age != 10 {
		t.Errorf("Age = %d, want 10 (incremented, not reset by a flower roll)", c.Age)
	}
}
