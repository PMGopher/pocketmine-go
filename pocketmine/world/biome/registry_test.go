package biome

import "testing"

func TestNewRegistryRegistersEveryVanillaBiomeAtItsRealID(t *testing.T) {
	r := NewRegistry()

	cases := []struct {
		id   int
		name string
	}{
		{IDOcean, "Ocean"},
		{IDPlains, "Plains"},
		{IDDesert, "Desert"},
		{IDExtremeHills, "Mountains"},
		{IDForest, "Oak Forest"},
		{IDTaiga, "Taiga"},
		{IDSwampland, "Swamp"},
		{IDRiver, "River"},
		{IDHell, "Hell"},
		{IDIcePlains, "Ice Plains"},
		{IDExtremeHillsEdge, "Small Mountains"},
		{IDBirchForest, "Birch Forest"},
	}

	for _, c := range cases {
		b := r.GetBiome(c.id)
		if b.ID() != c.id {
			t.Errorf("GetBiome(%d).ID() = %d, want %d", c.id, b.ID(), c.id)
		}
		if b.Name() != c.name {
			t.Errorf("GetBiome(%d).Name() = %q, want %q", c.id, b.Name(), c.name)
		}
	}
}

func TestGetBiomeAutoRegistersUnknownBiomeForAnUnclaimedID(t *testing.T) {
	r := NewRegistry()

	const unclaimedID = 200
	b := r.GetBiome(unclaimedID)
	if b.Name() != "Unknown" {
		t.Errorf("GetBiome(%d).Name() = %q, want %q", unclaimedID, b.Name(), "Unknown")
	}
	if b.ID() != unclaimedID {
		t.Errorf("GetBiome(%d).ID() = %d, want %d", unclaimedID, b.ID(), unclaimedID)
	}

	// A second lookup must return the same registered instance, not register a fresh one each time.
	if again := r.GetBiome(unclaimedID); again != b {
		t.Error("expected a second GetBiome call for the same unclaimed ID to return the same instance")
	}
}

func TestGrassyBiomesShareGroundCoverButNotPopulatorSlices(t *testing.T) {
	plains := NewPlainBiome()
	swamp := NewSwampBiome()

	if len(plains.GroundCover()) != 5 {
		t.Fatalf("Plains ground cover length = %d, want 5", len(plains.GroundCover()))
	}
	if plains.GroundCover()[0].GetTypeId() != swamp.GroundCover()[0].GetTypeId() {
		t.Error("expected Plains and Swamp (both GrassyBiome) to have the same top ground cover block type")
	}

	if len(plains.Populators()) == 0 {
		t.Error("expected Plains to have a TallGrass populator")
	}
	if len(swamp.Populators()) != 0 {
		t.Error("expected Swamp to have no populators, matching real SwampBiome")
	}
}

func TestMountainsAndSmallMountainsHaveIndependentPopulatorInstances(t *testing.T) {
	a := NewMountainsBiome()
	b := NewSmallMountainsBiome()

	if len(a.Populators()) == 0 || len(b.Populators()) == 0 {
		t.Fatal("expected both Mountains variants to have populators")
	}
	if &a.Populators()[0] == &b.Populators()[0] {
		t.Error("expected NewSmallMountainsBiome to build its own fresh populator instances, not share Mountains'")
	}
	if a.MaxElevation() == b.MaxElevation() {
		t.Error("expected Small Mountains to have a different (lower) max elevation than Mountains")
	}
}
