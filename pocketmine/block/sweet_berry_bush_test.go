package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func newTestSweetBerryBush(w World) *SweetBerryBush {
	s := NewSweetBerryBush(mustBlockIdentifier(1089), "Test Sweet Berry Bush", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestSweetBerryBushOnEntityInsideDamagesLivingWhenGrown(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSweetBerryBush(w)
	s.Age = SweetBerryBushStageBushNoBerries

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	living.SetFallDistance(5)
	startHealth := living.GetHealth()

	if !s.OnEntityInside(living) {
		t.Fatal("expected OnEntityInside to return true")
	}
	if living.GetHealth() != startHealth-1 {
		t.Errorf("GetHealth() = %v, want %v", living.GetHealth(), startHealth-1)
	}
	if living.GetFallDistance() != 0 {
		t.Errorf("GetFallDistance() = %v, want 0", living.GetFallDistance())
	}
}

func TestSweetBerryBushOnEntityInsideSparesLivingWhileSapling(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSweetBerryBush(w)
	s.Age = SweetBerryBushStageSapling

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := living.GetHealth()

	s.OnEntityInside(living)

	if living.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v (a sapling shouldn't damage entities)", living.GetHealth(), startHealth)
	}
}
