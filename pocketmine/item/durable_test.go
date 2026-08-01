package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

// Compile-time proof that a real Durable-embedding item satisfies block.Durable (the
// forward-compatible marker in block/candle_component.go).
var (
	_ Item          = (*FlintSteel)(nil)
	_ block.Item    = (*FlintSteel)(nil)
	_ block.Durable = (*FlintSteel)(nil)
)

func newTestFlintSteel() *FlintSteel {
	return NewFlintSteel(NewItemIdentifier(FLINT_AND_STEEL), "Flint and Steel")
}

func TestFlintSteelGetMaxDurability(t *testing.T) {
	f := newTestFlintSteel()
	if f.GetMaxDurability() != 65 {
		t.Errorf("GetMaxDurability() = %d, want 65", f.GetMaxDurability())
	}
}

func TestDurableApplyDamageIncreasesDamage(t *testing.T) {
	f := newTestFlintSteel()
	if !f.ApplyDamage(10) {
		t.Fatal("expected ApplyDamage to succeed")
	}
	if f.GetDamage() != 10 {
		t.Errorf("GetDamage() = %d, want 10", f.GetDamage())
	}
}

// TestDurableApplyDamageExceedingMaxDurabilityBreaksTheItem exercises the min(...) cap indirectly:
// damage exceeding max durability is capped at 65, which immediately trips IsBroken() inside the
// same ApplyDamage call, triggering onBroken() (pop + damage reset to 0) before it returns - so
// GetDamage() afterward is observably 0, not 65. See TestDurableOnBrokenPopsAndResetsDamage for
// the same mechanism tested directly against IsBroken/onBroken.
func TestDurableApplyDamageExceedingMaxDurabilityBreaksTheItem(t *testing.T) {
	f := newTestFlintSteel()
	f.SetCount(2)
	f.ApplyDamage(1000)
	if f.GetDamage() != 0 {
		t.Errorf("GetDamage() = %d, want 0 (breaking resets damage)", f.GetDamage())
	}
	if f.GetCount() != 1 {
		t.Errorf("GetCount() = %d, want 1 (breaking pops one)", f.GetCount())
	}
}

func TestDurableUnbreakableRejectsDamage(t *testing.T) {
	f := newTestFlintSteel()
	f.SetUnbreakable(true)
	if f.ApplyDamage(10) {
		t.Error("expected ApplyDamage to fail on an unbreakable item")
	}
	if f.GetDamage() != 0 {
		t.Errorf("GetDamage() = %d, want 0", f.GetDamage())
	}
}

func TestDurableOnBrokenPopsAndResetsDamage(t *testing.T) {
	f := newTestFlintSteel()
	f.SetCount(3)
	f.ApplyDamage(65) // exactly at max durability -> broken

	if f.GetCount() != 2 {
		t.Errorf("GetCount() = %d, want 2 (onBroken should pop one)", f.GetCount())
	}
	if f.GetDamage() != 0 {
		t.Errorf("GetDamage() = %d, want 0 (onBroken should reset damage)", f.GetDamage())
	}
}

func TestDurableIsBrokenWhenNull(t *testing.T) {
	f := newTestFlintSteel()
	f.SetCount(0)
	if !f.IsBroken() {
		t.Error("expected a null item stack to be considered broken")
	}
}

func TestDurableSetDamageRejectsOutOfRange(t *testing.T) {
	f := newTestFlintSteel()
	defer func() {
		if recover() == nil {
			t.Error("expected SetDamage to panic for an out-of-range value")
		}
	}()
	f.SetDamage(1000)
}

func TestDurableNbtRoundTrip(t *testing.T) {
	f := newTestFlintSteel()
	f.ApplyDamage(20)
	f.SetUnbreakable(true)

	decoded := newTestFlintSteel()
	decoded.SetNamedTag(f.GetNamedTag())

	if decoded.GetDamage() != 20 {
		t.Errorf("GetDamage() = %d, want 20", decoded.GetDamage())
	}
	if !decoded.IsUnbreakable() {
		t.Error("expected IsUnbreakable() to round trip as true")
	}
}

func TestToolGetMaxStackSize(t *testing.T) {
	f := newTestFlintSteel()
	if f.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", f.GetMaxStackSize())
	}
}

func TestToolGetMiningEfficiency(t *testing.T) {
	f := newTestFlintSteel()
	if eff := f.GetMiningEfficiency(true); eff != 1 {
		t.Errorf("GetMiningEfficiency(true) = %v, want 1", eff)
	}
	if eff := f.GetMiningEfficiency(false); eff != 1 {
		t.Errorf("GetMiningEfficiency(false) = %v, want 1", eff)
	}
}
