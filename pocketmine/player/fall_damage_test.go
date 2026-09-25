package player

import (
	"testing"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
)

// groundY returns the Y of the top surface of the test world's terrain at x=0,z=0 (the feet
// position of a player standing on it).
func groundY(t *testing.T, w *world.World) float64 {
	t.Helper()
	chunk := w.GetOrLoadChunk(0, 0)
	h, ok := chunk.GetHighestBlockAt(0, 0)
	if !ok {
		t.Fatal("test world has no terrain at 0,0")
	}
	return float64(h + 1)
}

// fallTo moves p down one block per movement (like per-tick PlayerAuthInput reports) until its
// feet reach y.
func fallTo(t *testing.T, p *Player, y float64) {
	t.Helper()
	for pos := p.GetPosition(); pos.Y > y; pos = p.GetPosition() {
		next := max(pos.Y-1, y)
		p.HandleMovement(math.NewVector3(pos.X, next, pos.Z))
		if p.GetPosition().Y != next {
			t.Fatalf("HandleMovement refused a 1-block move to y=%v", next)
		}
	}
}

func TestShortFallDealsNoDamage(t *testing.T) {
	w := newTestWorld(t)
	ground := groundY(t, w)
	p := newTestPlayerIn(t, w, math.NewVector3(0.5, ground+2, 0.5))

	fallTo(t, p, ground)

	if p.GetHealth() != float64(p.GetMaxHealth()) {
		t.Errorf("GetHealth() = %v, want unchanged %v after a 2-block fall", p.GetHealth(), p.GetMaxHealth())
	}
	if !p.IsOnGround() {
		t.Error("IsOnGround() = false after landing, want true")
	}
}

func TestLongFallDealsLivingFallDamage(t *testing.T) {
	w := newTestWorld(t)
	ground := groundY(t, w)
	p := newTestPlayerIn(t, w, math.NewVector3(0.5, ground+10, 0.5))
	startHealth := p.GetHealth()

	fallTo(t, p, ground)

	want := startHealth - p.CalculateFallDamage(10)
	if p.GetHealth() != want {
		t.Errorf("GetHealth() = %v, want %v after a 10-block fall", p.GetHealth(), want)
	}
	if p.GetFallDistance() != 0 {
		t.Errorf("GetFallDistance() = %v, want 0 after landing", p.GetFallDistance())
	}
}

func TestFlyingPlayersTakeNoFallDamage(t *testing.T) {
	w := newTestWorld(t)
	ground := groundY(t, w)
	p := newTestPlayerIn(t, w, math.NewVector3(0.5, ground+10, 0.5))
	p.SetFlying(true)

	fallTo(t, p, ground)

	if p.GetHealth() != float64(p.GetMaxHealth()) {
		t.Errorf("GetHealth() = %v, want unchanged (Player::calculateFallDamage is 0 while flying)", p.GetHealth())
	}
}

func TestHandleMovementRefusesMovesOfMoreThan15Blocks(t *testing.T) {
	w := newTestWorld(t)
	ground := groundY(t, w)
	p := newTestPlayerIn(t, w, math.NewVector3(0.5, ground, 0.5))

	p.HandleMovement(math.NewVector3(0.5, ground, 20.5))
	if p.GetPosition().Z != 0.5 {
		t.Errorf("position Z = %v after a refused move, want unchanged 0.5", p.GetPosition().Z)
	}
}
