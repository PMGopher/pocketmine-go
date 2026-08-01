package item

import "testing"

var (
	_ Item = (*BakedPotato)(nil)
	_ Item = (*Beetroot)(nil)
	_ Item = (*CookedChicken)(nil)
	_ Item = (*CookedFish)(nil)
	_ Item = (*CookedMutton)(nil)
	_ Item = (*CookedPorkchop)(nil)
	_ Item = (*CookedRabbit)(nil)
	_ Item = (*CookedSalmon)(nil)
	_ Item = (*GoldenApple)(nil)
	_ Item = (*GoldenCarrot)(nil)
	_ Item = (*MushroomStew)(nil)
	_ Item = (*PoisonousPotato)(nil)
	_ Item = (*PumpkinPie)(nil)
	_ Item = (*RabbitStew)(nil)
	_ Item = (*RawBeef)(nil)
	_ Item = (*RawChicken)(nil)
	_ Item = (*RawFish)(nil)
	_ Item = (*RawMutton)(nil)
	_ Item = (*RawPorkchop)(nil)
	_ Item = (*RawRabbit)(nil)
	_ Item = (*RawSalmon)(nil)
	_ Item = (*RottenFlesh)(nil)
	_ Item = (*Steak)(nil)
	_ Item = (*GlowBerries)(nil)
	_ Item = (*SweetBerries)(nil)
	_ Item = (*DriedKelp)(nil)
	_ Item = (*Cookie)(nil)
	_ Item = (*Melon)(nil)
	_ Item = (*BeetrootSoup)(nil)
)

func TestSimpleFoodValues(t *testing.T) {
	cases := []struct {
		name                  string
		food                  Item
		wantFoodRestore       int
		wantSaturationRestore float64
	}{
		{"BakedPotato", NewBakedPotato(NewItemIdentifier(BAKED_POTATO), "BakedPotato"), 5, 7.2},
		{"Beetroot", NewBeetroot(NewItemIdentifier(BEETROOT), "Beetroot"), 1, 1.2},
		{"CookedChicken", NewCookedChicken(NewItemIdentifier(COOKED_CHICKEN), "CookedChicken"), 6, 7.2},
		{"CookedFish", NewCookedFish(NewItemIdentifier(COOKED_FISH), "CookedFish"), 5, 6},
		{"CookedMutton", NewCookedMutton(NewItemIdentifier(COOKED_MUTTON), "CookedMutton"), 6, 9.6},
		{"CookedPorkchop", NewCookedPorkchop(NewItemIdentifier(COOKED_PORKCHOP), "CookedPorkchop"), 8, 12.8},
		{"CookedRabbit", NewCookedRabbit(NewItemIdentifier(COOKED_RABBIT), "CookedRabbit"), 5, 6},
		{"CookedSalmon", NewCookedSalmon(NewItemIdentifier(COOKED_SALMON), "CookedSalmon"), 6, 9.6},
		{"GoldenCarrot", NewGoldenCarrot(NewItemIdentifier(GOLDEN_CARROT), "GoldenCarrot"), 6, 14.4},
		{"PumpkinPie", NewPumpkinPie(NewItemIdentifier(PUMPKIN_PIE), "PumpkinPie"), 8, 4.8},
		{"RawBeef", NewRawBeef(NewItemIdentifier(RAW_BEEF), "RawBeef"), 3, 1.8},
		{"RawChicken", NewRawChicken(NewItemIdentifier(RAW_CHICKEN), "RawChicken"), 2, 1.2},
		{"RawFish", NewRawFish(NewItemIdentifier(RAW_FISH), "RawFish"), 2, 0.4},
		{"RawMutton", NewRawMutton(NewItemIdentifier(RAW_MUTTON), "RawMutton"), 2, 1.2},
		{"RawPorkchop", NewRawPorkchop(NewItemIdentifier(RAW_PORKCHOP), "RawPorkchop"), 3, 0.6},
		{"RawRabbit", NewRawRabbit(NewItemIdentifier(RAW_RABBIT), "RawRabbit"), 3, 1.8},
		{"RawSalmon", NewRawSalmon(NewItemIdentifier(RAW_SALMON), "RawSalmon"), 2, 0.2},
		{"RottenFlesh", NewRottenFlesh(NewItemIdentifier(ROTTEN_FLESH), "RottenFlesh"), 4, 0.8},
		{"Steak", NewSteak(NewItemIdentifier(STEAK), "Steak"), 8, 12.8},
		{"GlowBerries", NewGlowBerries(NewItemIdentifier(GLOW_BERRIES), "GlowBerries"), 2, 0.4},
		{"SweetBerries", NewSweetBerries(NewItemIdentifier(SWEET_BERRIES), "SweetBerries"), 2, 1.2},
		{"DriedKelp", NewDriedKelp(NewItemIdentifier(DRIED_KELP), "DriedKelp"), 1, 0.6},
		{"Cookie", NewCookie(NewItemIdentifier(COOKIE), "Cookie"), 2, 0.4},
		{"Melon", NewMelon(NewItemIdentifier(MELON), "Melon"), 2, 1.2},
		{"PoisonousPotato", NewPoisonousPotato(NewItemIdentifier(POISONOUS_POTATO), "PoisonousPotato"), 2, 1.2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			restorer := c.food.(interface{ GetFoodRestore() int })
			if got := restorer.GetFoodRestore(); got != c.wantFoodRestore {
				t.Errorf("GetFoodRestore() = %d, want %d", got, c.wantFoodRestore)
			}
			saturator := c.food.(interface{ GetSaturationRestore() float64 })
			if got := saturator.GetSaturationRestore(); got != c.wantSaturationRestore {
				t.Errorf("GetSaturationRestore() = %v, want %v", got, c.wantSaturationRestore)
			}
		})
	}
}

func TestGoldenAppleDoesNotRequireHunger(t *testing.T) {
	g := NewGoldenApple(NewItemIdentifier(GOLDEN_APPLE), "Golden Apple")
	if g.RequiresHunger() {
		t.Error("expected GoldenApple.RequiresHunger() to be false")
	}
}

func TestStewsHaveMaxStackSizeOne(t *testing.T) {
	stews := []interface{ GetMaxStackSize() int }{
		NewMushroomStew(NewItemIdentifier(MUSHROOM_STEW), "Mushroom Stew"),
		NewRabbitStew(NewItemIdentifier(RABBIT_STEW), "Rabbit Stew"),
		NewBeetrootSoup(NewItemIdentifier(BEETROOT_SOUP), "Beetroot Soup"),
	}
	for _, s := range stews {
		if s.GetMaxStackSize() != 1 {
			t.Errorf("GetMaxStackSize() = %d, want 1", s.GetMaxStackSize())
		}
	}
}
