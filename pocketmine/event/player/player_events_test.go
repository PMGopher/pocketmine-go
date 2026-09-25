package player

import (
	"testing"

	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
)

type testPlayer struct{ name string }

func (p *testPlayer) GetID() int                { return 1 }
func (p *testPlayer) GetPosition() math.Vector3 { return math.Vector3{} }
func (p *testPlayer) IsClosed() bool            { return false }
func (p *testPlayer) GetName() string           { return p.name }
func (p *testPlayer) GetDisplayName() string    { return p.name }

func TestPreLoginFinalReasonFollowsPriority(t *testing.T) {
	ev := NewPlayerPreLoginEvent(nil, "127.0.0.1", 19132, true)
	if !ev.IsAllowed() || ev.GetFinalDisconnectReason() != "" {
		t.Fatalf("a fresh event must be allowed")
	}
	ev.SetKickFlag(KickFlagBanned, "banned", nil)
	ev.SetKickFlag(KickFlagServerFull, "full", "screen full")
	if ev.GetFinalDisconnectReason() != "full" || ev.GetFinalDisconnectScreenMessage() != "screen full" {
		t.Fatalf("server full must win over banned: %v", ev.GetFinalDisconnectReason())
	}
	if ev.GetDisconnectScreenMessage(KickFlagBanned) != "banned" {
		t.Fatalf("nil screen message must default to the reason")
	}
	ev.ClearKickFlag(KickFlagServerFull)
	if got := ev.GetKickFlags(); len(got) != 1 || got[0] != KickFlagBanned {
		t.Fatalf("kick flags = %v", got)
	}
}

func TestDeriveMessage(t *testing.T) {
	if got := DeriveMessage("Steve", nil); got.Text() != lang.KeyDeathAttackGeneric {
		t.Errorf("nil cause: %q", got.Text())
	}
	void := entityevent.NewEntityDamageEvent(&testPlayer{"Steve"}, entityevent.CauseVoid, 10, nil)
	if got := DeriveMessage("Steve", void); got.Text() != lang.KeyDeathAttackOutOfWorld {
		t.Errorf("void: %q", got.Text())
	}
	fall := entityevent.NewEntityDamageEvent(&testPlayer{"Steve"}, entityevent.CauseFall, 5, nil)
	if got := DeriveMessage("Steve", fall); got.Text() != lang.KeyDeathFellAccidentGeneric {
		t.Errorf("big fall: %q", got.Text())
	}

	DeathMessageClassifier.IsPlayer = func(e Entity) bool { _, ok := e.(*testPlayer); return ok }
	defer func() { DeathMessageClassifier.IsPlayer = nil }()
	attack := entityevent.NewEntityDamageByEntityEvent(&testPlayer{"Alex"}, &testPlayer{"Steve"}, entityevent.CauseEntityAttack, 3, nil)
	got := DeriveMessage("Steve", attack)
	if got.Text() != lang.KeyDeathAttackPlayer || got.Parameter(1) != "Alex" {
		t.Errorf("player attack: %q %v", got.Text(), got.Parameters())
	}
}

func TestPlayerDeathEventReachesEntityDeathHandlers(t *testing.T) {
	m := event.NewManager()
	called := false
	event.RegisterListener[entityevent.EntityDeathEvent](m, "p", event.Normal, false, func(e *entityevent.EntityDeathEvent) {
		called = true
		e.SetXpDropAmount(0)
	})
	ev := NewPlayerDeathEvent(&testPlayer{"Steve"}, nil, 7, "died", nil)
	event.CallOn(m, ev)
	if !called || ev.GetXpDropAmount() != 0 {
		t.Fatalf("EntityDeathEvent handler: called=%v xp=%d", called, ev.GetXpDropAmount())
	}
}
