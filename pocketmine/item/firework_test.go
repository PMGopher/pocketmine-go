package item

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/world/sound"
)

var (
	_ Item = (*FireworkStar)(nil)
	_ Item = (*FireworkRocket)(nil)
)

func TestFireworkRocketTypeExplosionSound(t *testing.T) {
	if _, ok := FireworkRocketTypeLargeBall.GetExplosionSound().(sound.FireworkLargeExplosionSound); !ok {
		t.Error("expected LargeBall to use FireworkLargeExplosionSound")
	}
	if _, ok := FireworkRocketTypeStar.GetExplosionSound().(sound.FireworkExplosionSound); !ok {
		t.Error("expected Star to use FireworkExplosionSound")
	}
}

func TestFireworkStarDefaultsToBlack(t *testing.T) {
	f := NewFireworkStar(NewItemIdentifier(FIREWORK_STAR), "Firework Star")
	if len(f.GetExplosion().GetColors()) != 1 || f.GetExplosion().GetColors()[0] != blockutils.DyeColorBlack {
		t.Errorf("GetExplosion().GetColors() = %v, want [Black]", f.GetExplosion().GetColors())
	}
	if _, ok := f.GetCustomColor(); ok {
		t.Error("expected a fresh FireworkStar not to have a custom color")
	}
}

func TestFireworkRocketExplosionRequiresNonEmptyColors(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected NewFireworkRocketExplosion to panic on empty colors")
		}
	}()
	NewFireworkRocketExplosion(FireworkRocketTypeStar, nil, nil, false, false)
}

func TestFireworkRocketExplosionColorMixAndFlash(t *testing.T) {
	e := NewFireworkRocketExplosion(FireworkRocketTypeBurst, []blockutils.DyeColor{blockutils.DyeColorRed, blockutils.DyeColorBlue}, nil, true, true)
	if e.GetFlashColor() != blockutils.DyeColorRed {
		t.Errorf("GetFlashColor() = %v, want Red (first color)", e.GetFlashColor())
	}
	mix := e.GetColorMix()
	want := color.Mix(blockutils.DyeColorRed.GetRgbValue(), blockutils.DyeColorBlue.GetRgbValue())
	if !mix.Equals(want) {
		t.Errorf("GetColorMix() = %+v, want %+v", mix, want)
	}
	if !e.WillTwinkle() || !e.GetTrail() {
		t.Error("expected WillTwinkle() and GetTrail() to both be true")
	}
}

func TestFireworkRocketExplosionRoundTripsThroughNBT(t *testing.T) {
	e := NewFireworkRocketExplosion(
		FireworkRocketTypeCreeper,
		[]blockutils.DyeColor{blockutils.DyeColorLime, blockutils.DyeColorPurple},
		[]blockutils.DyeColor{blockutils.DyeColorWhite},
		true, false,
	)

	decoded, err := FireworkRocketExplosionFromCompoundTag(e.ToCompoundTag())
	if err != nil {
		t.Fatalf("FireworkRocketExplosionFromCompoundTag: %v", err)
	}
	if decoded.GetType() != FireworkRocketTypeCreeper {
		t.Errorf("GetType() = %v, want Creeper", decoded.GetType())
	}
	if len(decoded.GetColors()) != 2 || decoded.GetColors()[0] != blockutils.DyeColorLime || decoded.GetColors()[1] != blockutils.DyeColorPurple {
		t.Errorf("GetColors() = %v, want [Lime, Purple]", decoded.GetColors())
	}
	if len(decoded.GetFadeColors()) != 1 || decoded.GetFadeColors()[0] != blockutils.DyeColorWhite {
		t.Errorf("GetFadeColors() = %v, want [White]", decoded.GetFadeColors())
	}
	if !decoded.WillTwinkle() || decoded.GetTrail() {
		t.Errorf("WillTwinkle()=%v GetTrail()=%v, want true/false", decoded.WillTwinkle(), decoded.GetTrail())
	}
}

func TestFireworkStarCustomColorRoundTripsThroughNBT(t *testing.T) {
	f := NewFireworkStar(NewItemIdentifier(FIREWORK_STAR), "Firework Star")
	f.SetExplosion(NewFireworkRocketExplosion(FireworkRocketTypeStar, []blockutils.DyeColor{blockutils.DyeColorRed}, nil, false, false))
	f.SetCustomColor(blockutils.DyeColorBlue.GetRgbValue())

	decoded := NewFireworkStar(NewItemIdentifier(FIREWORK_STAR), "Firework Star")
	decoded.SetNamedTag(f.GetNamedTag())

	c, ok := decoded.GetCustomColor()
	if !ok {
		t.Fatal("expected the decoded item to have a custom color")
	}
	if !c.Equals(blockutils.DyeColorBlue.GetRgbValue()) {
		t.Errorf("GetCustomColor() = %+v, want Blue's RGB", c)
	}
	if len(decoded.GetExplosion().GetColors()) != 1 || decoded.GetExplosion().GetColors()[0] != blockutils.DyeColorRed {
		t.Errorf("GetExplosion().GetColors() = %v, want [Red]", decoded.GetExplosion().GetColors())
	}
}

func TestFireworkRocketFlightTimeMultiplierRange(t *testing.T) {
	f := NewFireworkRocket(NewItemIdentifier(FIREWORK_ROCKET), "Firework Rocket")
	if f.GetFlightTimeMultiplier() != 1 {
		t.Errorf("GetFlightTimeMultiplier() = %d, want 1", f.GetFlightTimeMultiplier())
	}
	f.SetFlightTimeMultiplier(5)
	if f.GetFlightTimeMultiplier() != 5 {
		t.Errorf("GetFlightTimeMultiplier() = %d, want 5", f.GetFlightTimeMultiplier())
	}

	defer func() {
		if recover() == nil {
			t.Error("expected SetFlightTimeMultiplier to panic for an out-of-range value")
		}
	}()
	f.SetFlightTimeMultiplier(200)
}

func TestFireworkRocketExplosionsRoundTripThroughNBT(t *testing.T) {
	f := NewFireworkRocket(NewItemIdentifier(FIREWORK_ROCKET), "Firework Rocket")
	f.SetFlightTimeMultiplier(3)
	f.SetExplosions([]FireworkRocketExplosion{
		NewFireworkRocketExplosion(FireworkRocketTypeStar, []blockutils.DyeColor{blockutils.DyeColorYellow}, nil, false, false),
		NewFireworkRocketExplosion(FireworkRocketTypeBurst, []blockutils.DyeColor{blockutils.DyeColorCyan}, nil, false, true),
	})

	decoded := NewFireworkRocket(NewItemIdentifier(FIREWORK_ROCKET), "Firework Rocket")
	decoded.SetNamedTag(f.GetNamedTag())

	if decoded.GetFlightTimeMultiplier() != 3 {
		t.Errorf("GetFlightTimeMultiplier() = %d, want 3", decoded.GetFlightTimeMultiplier())
	}
	explosions := decoded.GetExplosions()
	if len(explosions) != 2 {
		t.Fatalf("len(GetExplosions()) = %d, want 2", len(explosions))
	}
	if explosions[0].GetType() != FireworkRocketTypeStar || explosions[1].GetType() != FireworkRocketTypeBurst {
		t.Errorf("explosion types = [%v, %v], want [Star, Burst]", explosions[0].GetType(), explosions[1].GetType())
	}
}
