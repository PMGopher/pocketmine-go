package item

import "testing"

var (
	_ Item = (*Potion)(nil)
	_ Item = (*SplashPotion)(nil)
)

func TestPotionTypeGetDisplayName(t *testing.T) {
	if got := PotionTypeStrongTurtleMaster.GetDisplayName(); got != "Strong Turtle Master" {
		t.Errorf("GetDisplayName() = %q, want %q", got, "Strong Turtle Master")
	}
	if got := PotionTypeWater.GetDisplayName(); got != "Water" {
		t.Errorf("GetDisplayName() = %q, want %q", got, "Water")
	}
}

func TestPotionDefaultsToWater(t *testing.T) {
	p := NewPotion(NewItemIdentifier(POTION), "Potion")
	if p.GetType() != PotionTypeWater {
		t.Errorf("GetType() = %v, want Water", p.GetType())
	}
	if p.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", p.GetMaxStackSize())
	}
}

func TestPotionSetTypeAndStateIdChanges(t *testing.T) {
	p := NewPotion(NewItemIdentifier(POTION), "Potion")
	water := p.GetStateId()

	p.SetType(PotionTypeHealing)
	if p.GetType() != PotionTypeHealing {
		t.Errorf("GetType() = %v, want Healing", p.GetType())
	}
	if p.GetStateId() == water {
		t.Error("expected GetStateId() to change when the potion type changes")
	}
}

func TestSplashPotionLingerFlag(t *testing.T) {
	splash := NewSplashPotion(NewItemIdentifier(SPLASH_POTION), "Splash Potion", false)
	if splash.Linger {
		t.Error("expected a plain SplashPotion not to linger")
	}
	lingering := NewSplashPotion(NewItemIdentifier(LINGERING_POTION), "Lingering Potion", true)
	if !lingering.Linger {
		t.Error("expected a lingering SplashPotion to have Linger=true")
	}
	if lingering.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", lingering.GetMaxStackSize())
	}
}
