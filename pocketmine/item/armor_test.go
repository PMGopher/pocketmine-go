package item

import (
	"testing"

	"pocketmine-go/pocketmine/color"
)

var _ Item = (*Armor)(nil)

func newTestLeatherCap() *Armor {
	material := NewArmorMaterial(15, nil)
	info := NewArmorTypeInfo(1, 55, 0, material)
	return NewArmor(NewItemIdentifier(LEATHER_CAP), "Leather Cap", info)
}

func TestArmorDefensePointsAndDurabilityFollowTypeInfo(t *testing.T) {
	a := newTestLeatherCap()
	if a.GetDefensePoints() != 1 {
		t.Errorf("GetDefensePoints() = %d, want 1", a.GetDefensePoints())
	}
	if a.GetMaxDurability() != 55 {
		t.Errorf("GetMaxDurability() = %d, want 55", a.GetMaxDurability())
	}
	if a.GetArmorSlot() != 0 {
		t.Errorf("GetArmorSlot() = %d, want 0", a.GetArmorSlot())
	}
}

func TestArmorGetMaxStackSize(t *testing.T) {
	a := newTestLeatherCap()
	if a.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", a.GetMaxStackSize())
	}
}

func TestArmorEnchantabilityFollowsMaterial(t *testing.T) {
	a := newTestLeatherCap()
	if a.GetEnchantability() != 15 {
		t.Errorf("GetEnchantability() = %d, want 15", a.GetEnchantability())
	}
}

func TestArmorApplyDamageUsesTypeInfoDurability(t *testing.T) {
	a := newTestLeatherCap()
	if !a.ApplyDamage(10) {
		t.Fatal("expected ApplyDamage to succeed")
	}
	if a.GetDamage() != 10 {
		t.Errorf("GetDamage() = %d, want 10", a.GetDamage())
	}
}

func TestArmorCustomColorDefaultsToUnset(t *testing.T) {
	a := newTestLeatherCap()
	if _, ok := a.GetCustomColor(); ok {
		t.Error("expected a fresh Armor not to have a custom color")
	}
}

func TestArmorCustomColorRoundTripsThroughNBT(t *testing.T) {
	a := newTestLeatherCap()
	a.SetCustomColor(color.NewColor(10, 20, 30, 255))

	decoded := newTestLeatherCap()
	decoded.SetNamedTag(a.GetNamedTag())

	c, ok := decoded.GetCustomColor()
	if !ok {
		t.Fatal("expected the decoded item to have a custom color")
	}
	if c.GetR() != 10 || c.GetG() != 20 || c.GetB() != 30 {
		t.Errorf("GetCustomColor() = %+v, want R=10 G=20 B=30", c)
	}
}

func TestArmorClearCustomColor(t *testing.T) {
	a := newTestLeatherCap()
	a.SetCustomColor(color.NewColor(1, 2, 3, 255))
	a.ClearCustomColor()

	if _, ok := a.GetCustomColor(); ok {
		t.Error("expected ClearCustomColor to unset the custom color")
	}
}
