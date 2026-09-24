package effect

import "testing"

func TestEffectInstanceDefaults(t *testing.T) {
	inst := NewEffectInstance(VanillaSpeed())
	if inst.GetDuration() != VanillaSpeed().GetDefaultDuration() || inst.GetDuration() != 600 {
		t.Errorf("duration = %d, want the default 600", inst.GetDuration())
	}
	if inst.GetAmplifier() != 0 || inst.GetEffectLevel() != 1 {
		t.Errorf("amplifier/level = %d/%d, want 0/1", inst.GetAmplifier(), inst.GetEffectLevel())
	}
	if !inst.IsVisible() || inst.IsAmbient() {
		t.Error("a new instance should be visible and not ambient")
	}
}

func TestEffectInstanceDecreaseDuration(t *testing.T) {
	inst := NewEffectInstanceWith(VanillaPoison(), 10, 0)
	inst.DecreaseDuration(4)
	if inst.GetDuration() != 6 || inst.HasExpired() {
		t.Errorf("duration = %d (expired %v), want 6", inst.GetDuration(), inst.HasExpired())
	}
	inst.DecreaseDuration(100)
	if inst.GetDuration() != 0 || !inst.HasExpired() {
		t.Errorf("duration = %d, want clamped to 0 and expired", inst.GetDuration())
	}
}

func TestApplyIntervals(t *testing.T) {
	// PoisonEffect/RegenerationEffect/WitherEffect::canTick: 25/50/40 >> amplifier ticks.
	cases := []struct {
		e         Effect
		amplifier int
		interval  int
	}{
		{VanillaPoison(), 0, 25},
		{VanillaPoison(), 1, 12},
		{VanillaRegeneration(), 0, 50},
		{VanillaRegeneration(), 2, 12},
		{VanillaWither(), 0, 40},
	}
	for _, c := range cases {
		ticks := 0
		for d := 1000; d > 0; d-- {
			if c.e.CanTick(NewEffectInstanceWith(c.e, d, c.amplifier)) {
				ticks++
			}
		}
		if want := 1000 / c.interval; ticks != want {
			t.Errorf("%v amp %d ticks %d times in 1000 ticks, want %d", c.e.GetName(), c.amplifier, ticks, want)
		}
	}
	if !VanillaInstantHealth().CanTick(NewEffectInstanceWith(VanillaInstantHealth(), 1, 0)) {
		t.Error("instant effects should always tick")
	}
}

func TestBadEffects(t *testing.T) {
	for _, e := range []Effect{VanillaPoison(), VanillaWither(), VanillaSlowness(), VanillaInstantDamage()} {
		if !e.IsBad() {
			t.Errorf("%v.IsBad() = false", e.GetName())
		}
	}
	for _, e := range []Effect{VanillaSpeed(), VanillaRegeneration(), VanillaInstantHealth()} {
		if e.IsBad() {
			t.Errorf("%v.IsBad() = true", e.GetName())
		}
	}
}

func TestStringToEffectParser(t *testing.T) {
	got, ok := GetStringToEffectParser().Parse("fire_resistance")
	if !ok || got != VanillaFireResistance() {
		t.Errorf("Parse(fire_resistance) = (%v, %v), want the Fire Resistance singleton", got, ok)
	}
	if len(GetAllVanillaEffects()) != 27 {
		t.Errorf("GetAllVanillaEffects has %d effects, want 27 (VanillaEffectsInputs)", len(GetAllVanillaEffects()))
	}
}
