package block

import "testing"

func TestVanillaDirtReturnsIndependentClones(t *testing.T) {
	d1 := VanillaDirt()
	d2 := VanillaDirt()

	if d1 == d2 {
		t.Fatal("expected VanillaDirt() to return distinct instances, not the shared singleton")
	}
	if d1.GetTypeId() != DIRT {
		t.Errorf("GetTypeId() = %d, want DIRT (%d)", d1.GetTypeId(), DIRT)
	}
	if _, ok := d1.(*Dirt); !ok {
		t.Errorf("VanillaDirt() = %T, want *Dirt", d1)
	}
}

func TestVanillaAirTypeId(t *testing.T) {
	a := VanillaAir()
	if a.GetTypeId() != AIR {
		t.Errorf("GetTypeId() = %d, want AIR (%d)", a.GetTypeId(), AIR)
	}
}

func TestVanillaWaterTypeId(t *testing.T) {
	w := VanillaWater()
	if w.GetTypeId() != WATER {
		t.Errorf("GetTypeId() = %d, want WATER (%d)", w.GetTypeId(), WATER)
	}
}

func TestVanillaObsidianBreakInfo(t *testing.T) {
	o := VanillaObsidian()
	info := o.GetBreakInfo()
	if info.GetHardness() != 35.0 {
		t.Errorf("GetHardness() = %v, want 35.0", info.GetHardness())
	}
	if info.GetToolHarvestLevel() != 5 {
		t.Errorf("GetToolHarvestLevel() = %d, want 5 (diamond)", info.GetToolHarvestLevel())
	}
}

func TestVanillaNetherrackAndSoulSoilTypeIds(t *testing.T) {
	if got := VanillaNetherrack().GetTypeId(); got != NETHERRACK {
		t.Errorf("VanillaNetherrack().GetTypeId() = %d, want %d", got, NETHERRACK)
	}
	if got := VanillaSoulSoil().GetTypeId(); got != SOUL_SOIL {
		t.Errorf("VanillaSoulSoil().GetTypeId() = %d, want %d", got, SOUL_SOIL)
	}
}
