package player

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestAttackEntityDamagesAndKnocksBackTheTarget(t *testing.T) {
	w := newTestWorld(t)
	attacker := newTestPlayerIn(t, w, math.NewVector3(0, 70, 0))
	victim := newTestPlayerIn(t, w, math.NewVector3(1, 70, 0)) // within reach, east of the attacker
	startHealth := victim.GetHealth()

	if !attacker.AttackEntity(victim) {
		t.Fatal("AttackEntity() = false, want true for a reachable, non-spectator attack")
	}
	if victim.GetHealth() != startHealth-1 { // an empty hand deals 1 (Item::getAttackPoints)
		t.Errorf("GetHealth() = %v, want %v", victim.GetHealth(), startHealth-1)
	}
	if victim.GetMotion().X <= 0 {
		t.Errorf("GetMotion().X = %v, want > 0 (knocked back away from the attacker)", victim.GetMotion().X)
	}
}

func TestAttackEntityOnAFarAwayTargetReturnsFalseAndAppliesNoDamage(t *testing.T) {
	w := newTestWorld(t)
	attacker := newTestPlayerIn(t, w, math.NewVector3(0, 70, 0))
	victim := newTestPlayerIn(t, w, math.NewVector3(1000, 70, 0))
	startHealth := victim.GetHealth()

	if attacker.AttackEntity(victim) {
		t.Fatal("AttackEntity() = true, want false for an out-of-reach target")
	}
	if victim.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v", victim.GetHealth(), startHealth)
	}
}

func TestAttackEntityOnADeadTargetReturnsFalse(t *testing.T) {
	w := newTestWorld(t)
	attacker := newTestPlayerIn(t, w, math.NewVector3(0, 70, 0))
	victim := newTestPlayerIn(t, w, math.NewVector3(1, 70, 0))
	victim.SetHealth(0)

	if attacker.AttackEntity(victim) {
		t.Fatal("AttackEntity() = true, want false against an already-dead target")
	}
}

func TestAttackEntityFromASpectatorIsCancelled(t *testing.T) {
	w := newTestWorld(t)
	attacker := newTestPlayerIn(t, w, math.NewVector3(0, 70, 0))
	attacker.SetGamemode(GameModeSpectator)
	victim := newTestPlayerIn(t, w, math.NewVector3(1, 70, 0))
	startHealth := victim.GetHealth()

	if attacker.AttackEntity(victim) {
		t.Fatal("AttackEntity() = true, want false for a spectator attacker")
	}
	if victim.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v (spectator attack should be cancelled)", victim.GetHealth(), startHealth)
	}
}
