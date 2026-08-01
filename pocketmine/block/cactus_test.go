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
