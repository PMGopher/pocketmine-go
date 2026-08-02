package block

import "testing"

func newTestFarmlandAt(w *containerTileWorld, x, y, z int, wetness int) *Farmland {
	f := NewFarmland(mustBlockIdentifier(1091), "Test Farmland", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, x, y, z)
	f.Wetness = wetness
	w.blocks[[3]int{x, y, z}] = f
	return f
}

func TestCropGrowthCalculateMultiplierBaseIsOne(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCropsLeaf(w)

	if got := CropGrowthCalculateMultiplier(c); got != 1 {
		t.Errorf("CropGrowthCalculateMultiplier() = %v, want 1 (no farmland below)", got)
	}
}

func TestCropGrowthCalculateMultiplierHydratedFarmlandBonus(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCropsLeaf(w)
	newTestFarmlandAt(w, 1, 1, 3, 1) // directly below, hydrated

	if got := CropGrowthCalculateMultiplier(c); got != 4 { // 1 + ON_HYDRATED_FARMLAND_BONUS(3)
		t.Errorf("CropGrowthCalculateMultiplier() = %v, want 4", got)
	}
}

func TestCropGrowthCalculateMultiplierDryFarmlandBonus(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCropsLeaf(w)
	newTestFarmlandAt(w, 1, 1, 3, 0) // directly below, dry

	if got := CropGrowthCalculateMultiplier(c); got != 2 { // 1 + ON_DRY_FARMLAND_BONUS(1)
		t.Errorf("CropGrowthCalculateMultiplier() = %v, want 2", got)
	}
}

func TestCropGrowthCalculateMultiplierAdjacentHydratedFarmlandBonus(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCropsLeaf(w)
	newTestFarmlandAt(w, 2, 1, 3, 1) // adjacent (east), hydrated

	got := CropGrowthCalculateMultiplier(c)
	want := 1 + 3.0/4 // 1 + ADJACENT_HYDRATED_FARMLAND_BONUS
	if got != want {
		t.Errorf("CropGrowthCalculateMultiplier() = %v, want %v", got, want)
	}
}

func TestCropGrowthHasEnoughLight(t *testing.T) {
	w := &fakeWorld{} // GetPotentialLightAt returns 15 by default
	c := newTestCropsLeaf(w)

	if !CropGrowthHasEnoughLight(c) {
		t.Error("expected CropGrowthHasEnoughLight() to be true at light level 15")
	}
}

type lowPotentialLightWorld struct {
	fakeWorld
}

func (w *lowPotentialLightWorld) GetPotentialLightAt(x, y, z int) int { return 3 }

func TestCropGrowthHasEnoughLightFalseWhenDark(t *testing.T) {
	w := &lowPotentialLightWorld{}
	c := newTestCropsLeaf(w)

	if CropGrowthHasEnoughLight(c) {
		t.Error("expected CropGrowthHasEnoughLight() to be false at light level 3")
	}
}
