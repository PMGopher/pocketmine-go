package enchantment

import (
	"testing"

	entityevent "pocketmine-go/pocketmine/event/entity"
)

func TestProtectionFactor(t *testing.T) {
	// floor((6 + level^2) * typeModifier / 3), ProtectionEnchantment::getProtectionFactor.
	cases := []struct {
		e     Enchantment
		level int
		want  int
	}{
		{VanillaProtection(), 1, 1},
		{VanillaProtection(), 4, 5},
		{VanillaFeatherFalling(), 4, 18},
		{VanillaBlastProtection(), 2, 5},
	}
	for _, c := range cases {
		p := c.e.(*ProtectionEnchantment)
		if got := p.GetProtectionFactor(c.level); got != c.want {
			t.Errorf("%v level %d: GetProtectionFactor = %d, want %d", c.e.GetName(), c.level, got, c.want)
		}
	}
}

func TestProtectionIsApplicable(t *testing.T) {
	fall := entityevent.NewEntityDamageEvent(nil, entityevent.CauseFall, 1, nil)
	lava := entityevent.NewEntityDamageEvent(nil, entityevent.CauseLava, 1, nil)
	if !VanillaProtection().(*ProtectionEnchantment).IsApplicable(fall) {
		t.Error("Protection should apply to every cause")
	}
	if !VanillaFeatherFalling().(*ProtectionEnchantment).IsApplicable(fall) {
		t.Error("Feather Falling should apply to fall damage")
	}
	if VanillaFeatherFalling().(*ProtectionEnchantment).IsApplicable(lava) {
		t.Error("Feather Falling should not apply to lava")
	}
	if !VanillaFireProtection().(*ProtectionEnchantment).IsApplicable(lava) {
		t.Error("Fire Protection should apply to lava")
	}
}

func TestStringToEnchantmentParser(t *testing.T) {
	p := GetStringToEnchantmentParser()
	got, ok := p.Parse("Feather_Falling")
	if !ok || got != VanillaFeatherFalling() {
		t.Errorf("Parse(Feather_Falling) = (%v, %v), want the Feather Falling singleton", got, ok)
	}
	if _, ok := p.Parse("not_an_enchantment"); ok {
		t.Error("Parse accepted an unknown name")
	}
}

func TestEnchantmentInstance(t *testing.T) {
	inst := NewEnchantmentInstance(VanillaSharpness(), 3)
	if inst.GetType() != VanillaSharpness() || inst.GetLevel() != 3 {
		t.Errorf("instance = (%v, %d), want (Sharpness, 3)", inst.GetType(), inst.GetLevel())
	}
	if VanillaSharpness().GetMaxLevel() != 5 || VanillaMending().GetMaxLevel() != 1 {
		t.Error("max levels differ from VanillaEnchantments")
	}
}
