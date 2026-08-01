package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func TestLavaOnEntityInsideDamagesIgnitesAndResetsFallDistance(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLava(w)

	e := entity.NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	e.SetFallDistance(10)
	startHealth := e.GetHealth()

	if !l.OnEntityInside(e) {
		t.Fatal("expected OnEntityInside to return true")
	}
	if e.GetHealth() != startHealth-4 {
		t.Errorf("GetHealth() = %v, want %v", e.GetHealth(), startHealth-4)
	}
	if e.GetFireTicks() != 8*20 {
		t.Errorf("GetFireTicks() = %d, want %d", e.GetFireTicks(), 8*20)
	}
	if e.GetFallDistance() != 0 {
		t.Errorf("GetFallDistance() = %v, want 0", e.GetFallDistance())
	}
}
