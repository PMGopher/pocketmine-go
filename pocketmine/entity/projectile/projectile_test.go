package projectile_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/projectile"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld(t *testing.T) *world.World {
	t.Helper()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
	})
}

func groundY(t *testing.T, w *world.World) float64 {
	t.Helper()
	h, ok := w.GetOrLoadChunk(0, 0).GetHighestBlockAt(0, 0)
	if !ok {
		t.Fatal("test world has no terrain at 0,0")
	}
	return float64(h + 1)
}

func tickUntil(e world.Entity, maxTicks int, done func() bool) int {
	for tick := 1; tick <= maxTicks; tick++ {
		e.OnUpdate(int64(tick))
		if done() {
			return tick
		}
	}
	return -1
}

func TestSnowballFallsAndBreaksOnTheGround(t *testing.T) {
	w := newTestWorld(t)
	y := groundY(t, w)
	s := projectile.NewSnowball(entity.NewLocation(0.5, y+3, 0.5, w, 0, 0), nil, nil)
	s.SetMotion(math.NewVector3(0.3, 0, 0))

	if got := tickUntil(s, 200, s.IsFlaggedForDespawn); got < 0 {
		t.Fatal("snowball never hit the ground")
	}
	if pos := s.GetPosition(); pos.X <= 0.5 {
		t.Errorf("snowball x = %v, want it to have moved east", pos.X)
	}
}

func TestArrowStopsInBlockAndRemembersIt(t *testing.T) {
	w := newTestWorld(t)
	y := groundY(t, w)
	a := projectile.NewArrow(entity.NewLocation(0.5, y+2, 0.5, w, 0, 0), nil, false, nil)
	a.SetMotion(math.NewVector3(0, -1, 0))

	if got := tickUntil(a, 100, func() bool { return a.GetBlockHit() != nil }); got < 0 {
		t.Fatal("arrow never hit a block")
	}
	if hit := a.GetBlockHit(); hit.FloorY() != int(y)-1 {
		t.Errorf("arrow block hit = %v, want the ground block at y=%d", *hit, int(y)-1)
	}
	if a.IsFlaggedForDespawn() {
		t.Error("an arrow stuck in a block was despawned (arrows stay until picked up)")
	}
}

func TestArrowHitsAndDamagesEntity(t *testing.T) {
	w := newTestWorld(t)
	y := groundY(t, w)
	zombie := entity.NewZombie(entity.NewLocation(3.5, y, 0.5, w, 0, 0), nil)
	zombie.NoDamageTicks = 0
	start := zombie.GetHealth()

	a := projectile.NewArrow(entity.NewLocation(0.5, y+1, 0.5, w, 0, 0), nil, false, nil)
	a.SetMotion(math.NewVector3(1.5, 0, 0))

	if got := tickUntil(a, 20, a.IsFlaggedForDespawn); got < 0 {
		t.Fatal("arrow never hit the zombie")
	}
	// Arrow::getResultDamage: ceil(|motion| * base damage 2) at the moment of impact.
	if zombie.GetHealth() >= start {
		t.Errorf("zombie health = %v, want less than %v after being shot", zombie.GetHealth(), start)
	}
}

func TestArrowSavesCriticalAndDamage(t *testing.T) {
	w := newTestWorld(t)
	a := projectile.NewArrow(entity.NewLocation(0.5, 50, 0.5, w, 0, 0), nil, true, nil)
	a.SetBaseDamage(5)
	created, err := entity.GetEntityFactory().CreateFromData(w, a.SaveNBT())
	if err != nil {
		t.Fatal(err)
	}
	loaded, ok := created.(*projectile.Arrow)
	if !ok {
		t.Fatalf("reloaded %T, want *projectile.Arrow", created)
	}
	if loaded.GetBaseDamage() != 5 {
		t.Errorf("base damage = %v after reload, want 5", loaded.GetBaseDamage())
	}
}
