package item

import "testing"

func TestVanillaApplePropertiesAndIdentity(t *testing.T) {
	a := VanillaApple()
	if a.GetTypeId() != APPLE {
		t.Errorf("GetTypeId() = %d, want APPLE (%d)", a.GetTypeId(), APPLE)
	}
	if a.GetVanillaName() != "Apple" {
		t.Errorf("GetVanillaName() = %q, want %q", a.GetVanillaName(), "Apple")
	}
}

func TestVanillaGettersReturnIndependentClones(t *testing.T) {
	a1 := VanillaApple()
	a2 := VanillaApple()
	a1.SetCount(5)
	if a2.GetCount() == 5 {
		t.Error("mutating one VanillaApple() result affected another - singleton wasn't cloned")
	}
}

func TestVanillaCharcoalAndCoalAreDistinctTypeIDsSharingTheCoalStruct(t *testing.T) {
	coal := VanillaCoal()
	charcoal := VanillaCharcoal()
	if coal.GetTypeId() == charcoal.GetTypeId() {
		t.Error("VanillaCoal() and VanillaCharcoal() have the same type ID, want distinct (CHARCOAL != COAL)")
	}
	if coal.GetVanillaName() != "Coal" {
		t.Errorf("VanillaCoal().GetVanillaName() = %q, want \"Coal\"", coal.GetVanillaName())
	}
	if charcoal.GetVanillaName() != "Charcoal" {
		t.Errorf("VanillaCharcoal().GetVanillaName() = %q, want \"Charcoal\"", charcoal.GetVanillaName())
	}
}

func TestVanillaRecordsReturnsAllTwentyDiscsWithDistinctTypeIDs(t *testing.T) {
	records := VanillaRecords()
	if len(records) != 20 {
		t.Fatalf("VanillaRecords() returned %d entries, want 20", len(records))
	}
	seen := map[int]bool{}
	for _, r := range records {
		if seen[r.GetTypeId()] {
			t.Errorf("duplicate record type ID %d", r.GetTypeId())
		}
		seen[r.GetTypeId()] = true
	}
}

func TestVanillaWritableBookDisplayNameIsBookAndQuill(t *testing.T) {
	if got := VanillaWritableBook().GetVanillaName(); got != "Book & Quill" {
		t.Errorf("VanillaWritableBook().GetVanillaName() = %q, want \"Book & Quill\"", got)
	}
}

func TestVanillaLingeringAndSplashPotionAreDistinctTypeIDs(t *testing.T) {
	lingering := VanillaLingeringPotion()
	splash := VanillaSplashPotion()
	if lingering.GetTypeId() == splash.GetTypeId() {
		t.Error("VanillaLingeringPotion() and VanillaSplashPotion() have the same type ID, want distinct")
	}
}
