package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

func newTestCopperBulb(w World) *CopperBulb {
	idInfo, err := NewBlockIdentifier(1016, nil)
	if err != nil {
		panic(err)
	}
	c := NewCopperBulb(idInfo, "Test Copper Bulb", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCopperBulbTogglePoweredFlipsLitOnlyWhenTurningOn(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCopperBulb(w)

	c.TogglePowered(true)
	if !c.Powered || !c.Lit {
		t.Fatalf("Powered=%v Lit=%v, want both true after first toggle-on", c.Powered, c.Lit)
	}

	// Powering on again while already powered should be a no-op (same powered value).
	c.TogglePowered(true)
	if !c.Lit {
		t.Error("expected Lit to stay true when TogglePowered(true) is called while already powered")
	}

	c.TogglePowered(false)
	if c.Powered {
		t.Error("expected Powered to become false")
	}
	if !c.Lit {
		t.Error("expected Lit to be unchanged by powering off")
	}

	// Powering on again after being off should flip Lit again (it was true, now false).
	c.TogglePowered(true)
	if c.Lit {
		t.Error("expected a second power-on transition to flip Lit back off")
	}
}

func TestCopperBulbGetLightLevelScalesWithOxidationWhenLit(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCopperBulb(w)
	c.Lit = true

	cases := []struct {
		oxidation blockutils.CopperOxidation
		want      int
	}{
		{blockutils.CopperOxidationNone, 15},
		{blockutils.CopperOxidationExposed, 12},
		{blockutils.CopperOxidationWeathered, 8},
		{blockutils.CopperOxidationOxidized, 4},
	}
	for _, tc := range cases {
		c.Oxidation = tc.oxidation
		if got := c.GetLightLevel(); got != tc.want {
			t.Errorf("oxidation %v: GetLightLevel() = %d, want %d", tc.oxidation, got, tc.want)
		}
	}

	c.Lit = false
	if got := c.GetLightLevel(); got != 0 {
		t.Errorf("unlit: GetLightLevel() = %d, want 0", got)
	}
}
