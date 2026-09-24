package entity

import "testing"

func TestGetFinalDamageSumsModifiersAndClampsAtZero(t *testing.T) {
	ev := NewEntityDamageEvent(nil, CauseEntityAttack, 10, map[int]float64{ModifierStrength: 3})
	ev.SetModifier(-4, ModifierArmor)
	if got := ev.GetFinalDamage(); got != 9 {
		t.Errorf("GetFinalDamage() = %v, want 9", got)
	}
	ev.SetModifier(-100, ModifierResistance)
	if got := ev.GetFinalDamage(); got != 0 {
		t.Errorf("GetFinalDamage() = %v, want 0 (never negative)", got)
	}
	if ev.GetOriginalModifier(ModifierArmor) != 0 || ev.GetOriginalModifier(ModifierStrength) != 3 {
		t.Error("original modifiers changed after SetModifier")
	}
	if ev.GetAttackCooldown() != 10 {
		t.Errorf("GetAttackCooldown() = %d, want the default 10", ev.GetAttackCooldown())
	}
}

func TestCanBeReducedByArmor(t *testing.T) {
	for cause, want := range map[int]bool{
		CauseEntityAttack: true, CauseProjectile: true, CauseFire: true, CauseBlockExplosion: true,
		CauseFall: false, CauseFireTick: false, CauseSuffocation: false, CauseDrowning: false,
		CauseStarvation: false, CauseVoid: false, CauseMagic: false, CauseSuicide: false,
	} {
		if got := NewEntityDamageEvent(nil, cause, 1, nil).CanBeReducedByArmor(); got != want {
			t.Errorf("cause %d: CanBeReducedByArmor() = %v, want %v", cause, got, want)
		}
	}
}

func TestCancellation(t *testing.T) {
	var src DamageSource = NewEntityDamageEvent(nil, CauseFall, 1, nil)
	src.Cancel()
	if !src.IsCancelled() {
		t.Fatal("IsCancelled() = false after Cancel")
	}
	src.Uncancel()
	if src.IsCancelled() {
		t.Error("IsCancelled() = true after Uncancel")
	}
}
