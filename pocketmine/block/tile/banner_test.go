package tile

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestBannerSaveDataRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	b := NewBanner(w, math.Vector3{})
	b.SetBaseColor(blockutils.DyeColorRed)
	b.SetType(BannerTypeOminous)
	b.SetPatterns([]blockutils.BannerPatternLayer{
		blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeCreeper, blockutils.DyeColorWhite),
		blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeSkull, blockutils.DyeColorBlue),
	})

	saved := b.SaveNBT()

	decoded := NewBanner(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}

	if decoded.GetBaseColor() != blockutils.DyeColorRed {
		t.Errorf("GetBaseColor() = %v, want Red", decoded.GetBaseColor())
	}
	if decoded.GetType() != BannerTypeOminous {
		t.Errorf("GetType() = %d, want Ominous", decoded.GetType())
	}
	patterns := decoded.GetPatterns()
	if len(patterns) != 2 {
		t.Fatalf("len(patterns) = %d, want 2", len(patterns))
	}
	if patterns[0].GetType() != blockutils.BannerPatternTypeCreeper || patterns[0].GetColor() != blockutils.DyeColorWhite {
		t.Errorf("patterns[0] = %+v, want Creeper/White", patterns[0])
	}
	if patterns[1].GetType() != blockutils.BannerPatternTypeSkull || patterns[1].GetColor() != blockutils.DyeColorBlue {
		t.Errorf("patterns[1] = %+v, want Skull/Blue", patterns[1])
	}
}

func TestBannerReadSaveDataDefaultsToBlackOnMissingBaseColor(t *testing.T) {
	w := &fakeWorld{}
	b := NewBanner(w, math.Vector3{})
	b.SetBaseColor(blockutils.DyeColorRed) // should be overwritten by ReadSaveData's default

	if err := b.ReadSaveData(nbt.NewCompoundTag()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if b.GetBaseColor() != blockutils.DyeColorBlack {
		t.Errorf("GetBaseColor() = %v, want Black (default when Base tag is missing)", b.GetBaseColor())
	}
}

func TestDyeColorInvertedIDRoundTrip(t *testing.T) {
	for c := blockutils.DyeColorWhite; c <= blockutils.DyeColorBlack; c++ {
		id := blockutils.DyeColorToInvertedID(c)
		got, ok := blockutils.DyeColorFromInvertedID(id)
		if !ok || got != c {
			t.Errorf("color %v: inverted ID round trip = (%v, %v), want (%v, true)", c, got, ok, c)
		}
	}
}

func TestBannerPatternTypeIDRoundTrip(t *testing.T) {
	id, ok := blockutils.BannerPatternTypeToID(blockutils.BannerPatternTypeMojang)
	if !ok || id != "moj" {
		t.Fatalf("BannerPatternTypeToID(Mojang) = (%q, %v), want (\"moj\", true)", id, ok)
	}
	got, ok := blockutils.BannerPatternTypeFromID("moj")
	if !ok || got != blockutils.BannerPatternTypeMojang {
		t.Errorf("BannerPatternTypeFromID(\"moj\") = (%v, %v), want (Mojang, true)", got, ok)
	}
}

func TestHangingSignSaveID(t *testing.T) {
	w := &fakeWorld{}
	h := NewHangingSign(w, math.Vector3{})
	if h.SaveID() != "HangingSign" {
		t.Errorf("SaveID() = %q, want %q", h.SaveID(), "HangingSign")
	}
	// HangingSign embeds Sign, so it should still round-trip text correctly.
	h.SetText(blockutils.NewSignText([]string{"Hung"}, nil, false))
	saved := h.SaveNBT()
	decoded := NewHangingSign(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if lines := decoded.GetText().GetLines(); lines[0] != "Hung" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "Hung")
	}
}
