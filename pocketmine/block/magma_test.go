package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func newTestMagma(w World) *Magma {
	m := NewMagma(mustBlockIdentifier(1087), "Test Magma", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	m.SetPosition(w, 1, 2, 3)
	return m
}

func TestMagmaOnEntityInsideDamagesNonSneakingLiving(t *testing.T) {
	w := &fakeWorld{}
	m := newTestMagma(w)

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := living.GetHealth()

	if !m.OnEntityInside(living) {
		t.Fatal("expected OnEntityInside to return true")
	}
	if living.GetHealth() != startHealth-1 {
		t.Errorf("GetHealth() = %v, want %v", living.GetHealth(), startHealth-1)
	}
}

func TestMagmaOnEntityInsideSparesASneakingLiving(t *testing.T) {
	w := &fakeWorld{}
	m := newTestMagma(w)

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	living.SetSneaking(true)
	startHealth := living.GetHealth()

	m.OnEntityInside(living)

	if living.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v (sneaking should avoid magma damage)", living.GetHealth(), startHealth)
	}
}

func TestMagmaOnEntityInsideIgnoresNonLivingEntities(t *testing.T) {
	w := &fakeWorld{}
	m := newTestMagma(w)

	e := entity.NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := e.GetHealth()

	m.OnEntityInside(e)

	if e.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v (only Living entities take magma damage)", e.GetHealth(), startHealth)
	}
}
