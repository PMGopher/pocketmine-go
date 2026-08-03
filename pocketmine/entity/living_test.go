package entity

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestLivingAttackByEntityAppliesKnockbackAwayFromDamager(t *testing.T) {
	damager := NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	victim := newTestLiving()
	victim.SetPosition(math.NewVector3(1, 0, 0)) // east of the damager

	ev := NewEntityDamageByEntityEvent(damager, victim, EntityDamageCauseEntityAttack, 4, nil)
	victim.Attack(ev)

	if victim.GetHealth() != 20-4 {
		t.Errorf("GetHealth() = %v, want %v", victim.GetHealth(), 20-4.0)
	}

	motion := victim.GetMotion()
	if motion.X <= 0 {
		t.Errorf("GetMotion().X = %v, want > 0 (knocked back away from the damager to the east)", motion.X)
	}
	if motion.Y <= 0 {
		t.Errorf("GetMotion().Y = %v, want > 0 (knockback always has an upward component)", motion.Y)
	}
}

func TestLivingAttackByEntityCancelledAppliesNoKnockback(t *testing.T) {
	damager := NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	victim := newTestLiving()
	victim.SetPosition(math.NewVector3(1, 0, 0))

	ev := NewEntityDamageByEntityEvent(damager, victim, EntityDamageCauseEntityAttack, 4, nil)
	ev.Cancel()
	victim.Attack(ev)

	if victim.GetMotion() != (math.Vector3{}) {
		t.Errorf("GetMotion() = %v, want zero (attack was cancelled)", victim.GetMotion())
	}
}

func TestLivingKnockBackIsANoOpForAZeroVector(t *testing.T) {
	l := newTestLiving()
	l.KnockBack(0, 0, DefaultKnockbackForce, DefaultKnockbackVerticalLimit)
	if l.GetMotion() != (math.Vector3{}) {
		t.Errorf("GetMotion() = %v, want zero for a zero-length knockback vector", l.GetMotion())
	}
}

func TestLivingKnockBackClampsVerticalMotionToTheLimit(t *testing.T) {
	l := newTestLiving()
	l.SetMotion(math.NewVector3(0, 10, 0)) // pre-existing huge upward motion
	l.KnockBack(1, 0, DefaultKnockbackForce, DefaultKnockbackVerticalLimit)

	if l.GetMotion().Y > DefaultKnockbackVerticalLimit {
		t.Errorf("GetMotion().Y = %v, want <= %v (vertical knockback limit)", l.GetMotion().Y, DefaultKnockbackVerticalLimit)
	}
}
