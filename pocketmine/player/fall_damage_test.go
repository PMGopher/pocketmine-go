package player

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestTrackFallStateAppliesNoDamageForAShortFall(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))

	// Falling 2 blocks (less than the 3-block-free fall allowance) and landing.
	p.TrackFallState(69, false)
	p.TrackFallState(68, true)

	if p.GetHealth() != float64(p.GetMaxHealth()) {
		t.Errorf("GetHealth() = %v, want unchanged %v after a short fall", p.GetHealth(), p.GetMaxHealth())
	}
}

func TestTrackFallStateAppliesRealDamageForALongFall(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 90, 0))
	startHealth := p.GetHealth()

	// Fall 10 blocks then land - one TrackFallState call per block of descent, matching how
	// PlayerAuthInput reports position every tick.
	y := 90.0
	for i := 0; i < 10; i++ {
		y--
		p.TrackFallState(y, false)
	}
	p.TrackFallState(y, true) // land

	wantDamage := p.CalculateFallDamage(10)
	if p.GetHealth() != startHealth-wantDamage {
		t.Errorf("GetHealth() = %v, want %v (fall damage %v)", p.GetHealth(), startHealth-wantDamage, wantDamage)
	}
	if p.GetFallDistance() != 0 {
		t.Errorf("GetFallDistance() = %v, want 0 after landing", p.GetFallDistance())
	}
}

func TestTrackFallStateOnlyAppliesDamageOnceOnLanding(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 90, 0))

	y := 90.0
	for i := 0; i < 10; i++ {
		y--
		p.TrackFallState(y, false)
	}
	p.TrackFallState(y, true)
	healthAfterFirstLanding := p.GetHealth()

	// Reporting onGround=true again without an intervening fall shouldn't re-apply damage.
	p.TrackFallState(y, true)
	if p.GetHealth() != healthAfterFirstLanding {
		t.Errorf("GetHealth() = %v, want unchanged %v (damage should only apply once per landing)", p.GetHealth(), healthAfterFirstLanding)
	}
}
