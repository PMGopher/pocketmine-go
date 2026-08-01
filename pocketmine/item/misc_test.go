package item

import "testing"

var (
	_ Item = (*Coal)(nil)
	_ Item = (*Stick)(nil)
	_ Item = (*Bowl)(nil)
	_ Item = (*BlazeRod)(nil)
	_ Item = (*StringItem)(nil)
	_ Item = (*GlassBottle)(nil)
	_ Item = (*Compass)(nil)
	_ Item = (*Clock)(nil)
	_ Item = (*SpiderEye)(nil)
	_ Item = (*Potato)(nil)
	_ Item = (*GoldenAppleEnchanted)(nil)
	_ Item = (*TurtleHelmet)(nil)
)

func TestFuelTimes(t *testing.T) {
	cases := []struct {
		name string
		item interface{ GetFuelTime() int }
		want int
	}{
		{"Coal", NewCoal(NewItemIdentifier(COAL), "Coal"), 1600},
		{"Stick", NewStick(NewItemIdentifier(STICK), "Stick"), 100},
		{"Bowl", NewBowl(NewItemIdentifier(BOWL), "Bowl"), 200},
		{"BlazeRod", NewBlazeRod(NewItemIdentifier(BLAZE_ROD), "Blaze Rod"), 2400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.item.GetFuelTime(); got != c.want {
				t.Errorf("GetFuelTime() = %d, want %d", got, c.want)
			}
		})
	}
}

func TestSpiderEyeFoodValues(t *testing.T) {
	s := NewSpiderEye(NewItemIdentifier(SPIDER_EYE), "Spider Eye")
	if s.GetFoodRestore() != 2 {
		t.Errorf("GetFoodRestore() = %d, want 2", s.GetFoodRestore())
	}
	if s.GetSaturationRestore() != 3.2 {
		t.Errorf("GetSaturationRestore() = %v, want 3.2", s.GetSaturationRestore())
	}
}

func TestPotatoFoodValues(t *testing.T) {
	p := NewPotato(NewItemIdentifier(POTATO), "Potato")
	if p.GetFoodRestore() != 1 {
		t.Errorf("GetFoodRestore() = %d, want 1", p.GetFoodRestore())
	}
	if p.GetSaturationRestore() != 0.6 {
		t.Errorf("GetSaturationRestore() = %v, want 0.6", p.GetSaturationRestore())
	}
}

func TestGoldenAppleEnchantedInheritsGoldenAppleBehavior(t *testing.T) {
	g := NewGoldenAppleEnchanted(NewItemIdentifier(ENCHANTED_GOLDEN_APPLE), "Enchanted Golden Apple")
	if g.GetFoodRestore() != 4 {
		t.Errorf("GetFoodRestore() = %d, want 4 (inherited from GoldenApple)", g.GetFoodRestore())
	}
	if g.RequiresHunger() {
		t.Error("expected GoldenAppleEnchanted.RequiresHunger() to be false, inherited from GoldenApple")
	}
}

func TestTurtleHelmetInheritsArmorBehavior(t *testing.T) {
	material := NewArmorMaterial(9, nil)
	info := NewArmorTypeInfo(2, 275, 3, material)
	th := NewTurtleHelmet(NewItemIdentifier(TURTLE_HELMET), "Turtle Helmet", info)

	if th.GetDefensePoints() != 2 {
		t.Errorf("GetDefensePoints() = %d, want 2", th.GetDefensePoints())
	}
	if th.GetMaxDurability() != 275 {
		t.Errorf("GetMaxDurability() = %d, want 275", th.GetMaxDurability())
	}
	if !th.ApplyDamage(5) {
		t.Fatal("expected ApplyDamage to succeed")
	}
	if th.GetDamage() != 5 {
		t.Errorf("GetDamage() = %d, want 5", th.GetDamage())
	}
}
