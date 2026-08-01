package item

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

var (
	_ Item = (*Banner)(nil)
	_ Item = (*CoralFan)(nil)
)

func TestBannerDefaultsToBlack(t *testing.T) {
	b := NewBanner(NewItemIdentifier(BANNER), "Banner")
	if b.GetColor() != blockutils.DyeColorBlack {
		t.Errorf("GetColor() = %v, want Black", b.GetColor())
	}
	if b.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", b.GetFuelTime())
	}
}

func TestBannerPatternsRoundTripThroughNBT(t *testing.T) {
	b := NewBanner(NewItemIdentifier(BANNER), "Banner")
	b.SetColor(blockutils.DyeColorRed)
	b.SetPatterns([]blockutils.BannerPatternLayer{
		blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeCreeper, blockutils.DyeColorWhite),
		blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeSkull, blockutils.DyeColorBlue),
	})

	decoded := NewBanner(NewItemIdentifier(BANNER), "Banner")
	decoded.SetNamedTag(b.GetNamedTag())

	patterns := decoded.GetPatterns()
	if len(patterns) != 2 {
		t.Fatalf("len(GetPatterns()) = %d, want 2", len(patterns))
	}
	if patterns[0].GetType() != blockutils.BannerPatternTypeCreeper || patterns[0].GetColor() != blockutils.DyeColorWhite {
		t.Errorf("patterns[0] = %+v, want Creeper/White", patterns[0])
	}
	if patterns[1].GetType() != blockutils.BannerPatternTypeSkull || patterns[1].GetColor() != blockutils.DyeColorBlue {
		t.Errorf("patterns[1] = %+v, want Skull/Blue", patterns[1])
	}
}

func TestBannerStateIdChangesWithColor(t *testing.T) {
	b := NewBanner(NewItemIdentifier(BANNER), "Banner")
	black := b.GetStateId()
	b.SetColor(blockutils.DyeColorRed)
	if b.GetStateId() == black {
		t.Error("expected GetStateId() to change when the color changes")
	}
}

func TestBannerCloneDeepCopiesPatterns(t *testing.T) {
	b := NewBanner(NewItemIdentifier(BANNER), "Banner")
	b.SetPatterns([]blockutils.BannerPatternLayer{blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeCreeper, blockutils.DyeColorWhite)})

	clone := b.Clone().(*Banner)
	clone.SetPatterns(nil)

	if len(b.GetPatterns()) != 1 {
		t.Error("expected cloning not to affect the original banner's patterns")
	}
}

func TestCoralFanDefaultsAndSetters(t *testing.T) {
	c := NewCoralFan(NewItemIdentifier(CORAL_FAN), "Coral Fan")
	if c.GetCoralType() != blockutils.CoralTypeTube {
		t.Errorf("GetCoralType() = %v, want Tube (zero value)", c.GetCoralType())
	}
	if c.IsDead() {
		t.Error("expected a fresh CoralFan not to be dead")
	}

	c.SetCoralType(blockutils.CoralTypeFire)
	c.SetDead(true)
	if c.GetCoralType() != blockutils.CoralTypeFire || !c.IsDead() {
		t.Errorf("got CoralType=%v Dead=%v, want Fire/true", c.GetCoralType(), c.IsDead())
	}

	if c.GetMaxStackSize() != 64 {
		t.Errorf("GetMaxStackSize() = %d, want 64 (ItemBase default)", c.GetMaxStackSize())
	}
}

func TestCoralFanStateIdChangesWithType(t *testing.T) {
	c := NewCoralFan(NewItemIdentifier(CORAL_FAN), "Coral Fan")
	tube := c.GetStateId()
	c.SetDead(true)
	if c.GetStateId() == tube {
		t.Error("expected GetStateId() to change when Dead changes")
	}
}
