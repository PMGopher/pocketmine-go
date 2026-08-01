package entity

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// Compile-time proof that *Entity/*Living satisfy the local forward-compatible interfaces the
// block package already declared for a future entity type (see block/world.go).
var (
	_ block.Entity = (*Entity)(nil)
	_ block.Living = (*Living)(nil)
)

func newTestEntity() *Entity {
	return NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
}

func newTestLiving() *Living {
	return NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
}

func TestEntityFallDistanceGetSetReset(t *testing.T) {
	e := newTestEntity()
	e.SetFallDistance(5.5)
	if e.GetFallDistance() != 5.5 {
		t.Errorf("GetFallDistance() = %v, want 5.5", e.GetFallDistance())
	}
	e.ResetFallDistance()
	if e.GetFallDistance() != 0 {
		t.Errorf("GetFallDistance() = %v, want 0 after reset", e.GetFallDistance())
	}
}

func TestEntitySetOnFireOnlyIncreasesFireTicks(t *testing.T) {
	e := newTestEntity()
	e.SetOnFire(5) // 100 ticks
	if e.GetFireTicks() != 100 {
		t.Fatalf("GetFireTicks() = %d, want 100", e.GetFireTicks())
	}
	e.SetOnFire(2) // 40 ticks - less than current, should NOT decrease
	if e.GetFireTicks() != 100 {
		t.Errorf("GetFireTicks() = %d, want 100 (SetOnFire shouldn't decrease fire ticks)", e.GetFireTicks())
	}
	if !e.IsOnFire() {
		t.Error("expected IsOnFire() to be true")
	}
}

func TestEntitySetFireTicksBlockedWhenFireProof(t *testing.T) {
	f := &fireProofEntity{Entity: *newTestEntity()}
	f.Init(f)

	f.SetFireTicks(100)
	if f.GetFireTicks() != 0 {
		t.Errorf("GetFireTicks() = %d, want 0 (fire-proof entities can't be set on fire)", f.GetFireTicks())
	}
}

func TestEntitySetFireTicksPanicsOnNegative(t *testing.T) {
	e := newTestEntity()
	defer func() {
		if recover() == nil {
			t.Error("expected SetFireTicks(-1) to panic")
		}
	}()
	e.SetFireTicks(-1)
}

func TestEntityExtinguishClearsFireTicksAndFiresEvent(t *testing.T) {
	e := newTestEntity()
	e.SetFireTicks(100)

	var gotCause int
	handle := event.RegisterListener(event.Global(), nil, event.Normal, false, func(ev *EntityExtinguishEvent) {
		gotCause = ev.GetCause()
	})
	defer handle.Unregister()

	e.Extinguish()

	if e.IsOnFire() {
		t.Error("expected the entity not to be on fire after Extinguish")
	}
	if gotCause != EntityExtinguishCauseCustom {
		t.Errorf("EntityExtinguishEvent cause = %d, want CauseCustom", gotCause)
	}
}

func TestEntitySetMotionCanBeCancelled(t *testing.T) {
	e := newTestEntity()
	handle := event.RegisterListener(event.Global(), nil, event.Normal, false, func(ev *EntityMotionEvent) {
		ev.Cancel()
	})
	defer handle.Unregister()

	ok := e.SetMotion(math.NewVector3(1, 2, 3))
	if ok {
		t.Error("expected SetMotion to report false when the event is cancelled")
	}
	if e.GetMotion() != (math.Vector3{}) {
		t.Errorf("GetMotion() = %v, want zero (cancelled motion shouldn't apply)", e.GetMotion())
	}
}

func TestEntitySetHealthClampsToMaxHealth(t *testing.T) {
	e := newTestEntity()
	e.SetMaxHealth(20)
	e.SetHealth(999)
	if e.GetHealth() != 20 {
		t.Errorf("GetHealth() = %v, want 20 (clamped to max)", e.GetHealth())
	}
}

func TestEntitySetHealthToZeroKillsIt(t *testing.T) {
	e := newTestEntity()
	e.SetHealth(0)
	if e.IsAlive() {
		t.Error("expected the entity to be dead after SetHealth(0)")
	}
}

func TestEntityAttackReducesHealthByFinalDamage(t *testing.T) {
	e := newTestEntity()
	startHealth := e.GetHealth()

	ev := NewEntityDamageEvent(e, EntityDamageCauseEntityAttack, 4, nil)
	e.Attack(ev)

	if e.GetHealth() != startHealth-4 {
		t.Errorf("GetHealth() = %v, want %v", e.GetHealth(), startHealth-4)
	}
	if e.GetLastDamageCause() != ev {
		t.Error("expected SetLastDamageCause to record the event")
	}
}

func TestEntityAttackCancelledDoesNotReduceHealth(t *testing.T) {
	e := newTestEntity()
	startHealth := e.GetHealth()

	handle := event.RegisterListener(event.Global(), nil, event.Normal, false, func(ev *EntityDamageEvent) {
		ev.Cancel()
	})
	defer handle.Unregister()

	ev := NewEntityDamageEvent(e, EntityDamageCauseEntityAttack, 4, nil)
	e.Attack(ev)

	if e.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v (attack was cancelled)", e.GetHealth(), startHealth)
	}
}

func TestEntityAttackCancelsFireDamageWhenFireProof(t *testing.T) {
	f := &fireProofEntity{Entity: *newTestEntity()}
	f.Init(f)
	startHealth := f.GetHealth()

	ev := NewEntityDamageEvent(&f.Entity, EntityDamageCauseFire, 4, nil)
	f.Attack(ev)

	if !ev.IsCancelled() {
		t.Error("expected fire damage against a fire-proof entity to be cancelled")
	}
	if f.GetHealth() != startHealth {
		t.Errorf("GetHealth() = %v, want unchanged %v", f.GetHealth(), startHealth)
	}
}

func TestEntityDamageEventGetFinalDamageSumsModifiers(t *testing.T) {
	e := newTestEntity()
	ev := NewEntityDamageEvent(e, EntityDamageCauseFall, 10, map[int]float64{
		EntityDamageModifierArmor: -3,
	})
	if got := ev.GetFinalDamage(); got != 7 {
		t.Errorf("GetFinalDamage() = %v, want 7", got)
	}
}

func TestEntityDamageEventGetFinalDamageNeverNegative(t *testing.T) {
	e := newTestEntity()
	ev := NewEntityDamageEvent(e, EntityDamageCauseFall, 2, map[int]float64{
		EntityDamageModifierArmor: -10,
	})
	if got := ev.GetFinalDamage(); got != 0 {
		t.Errorf("GetFinalDamage() = %v, want 0 (clamped)", got)
	}
}

func TestLivingIsLivingAndSneaking(t *testing.T) {
	l := newTestLiving()
	if !l.IsLiving() {
		t.Error("expected IsLiving() to be true")
	}
	if l.IsSneaking() {
		t.Error("expected a fresh Living not to be sneaking")
	}
	l.SetSneaking(true)
	if !l.IsSneaking() {
		t.Error("expected IsSneaking() to be true after SetSneaking(true)")
	}
}

// fireProofEntity overrides IsFireProof via self-dispatch, exercising Entity's entityShaper
// pattern (mirrors block/tile/item's established self-dispatch conventions).
type fireProofEntity struct {
	Entity
}

func (f *fireProofEntity) IsFireProof() bool { return true }
