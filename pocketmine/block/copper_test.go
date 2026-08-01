package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

type fakeAxeItem struct {
	fakeItem
	damage int
}

func (a *fakeAxeItem) ApplyDamage(amount int) bool { a.damage += amount; return true }

func newTestCopper(w World) *Copper {
	idInfo, err := NewBlockIdentifier(1003, nil)
	if err != nil {
		panic(err)
	}
	c := NewCopper(idInfo, "Test Copper", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestCopperAxeScrapesOxidationThenRemovesWax(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCopper(w)
	c.SetOxidation(blockutils.CopperOxidationWeathered)
	c.SetWaxed(true)

	axe := &fakeAxeItem{}

	// Waxed: an axe should remove the wax first, without touching oxidation.
	if !c.OnInteract(axe, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle the axe interaction")
	}
	if c.IsWaxed() {
		t.Error("axe should have removed the wax")
	}
	if c.GetOxidation() != blockutils.CopperOxidationWeathered {
		t.Error("removing wax should not change oxidation")
	}
	if axe.damage != 1 {
		t.Errorf("axe damage = %d, want 1", axe.damage)
	}

	// Not waxed: an axe should scrape off one level of oxidation.
	if !c.OnInteract(axe, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle the axe interaction")
	}
	if c.GetOxidation() != blockutils.CopperOxidationExposed {
		t.Errorf("GetOxidation() = %v, want Exposed", c.GetOxidation())
	}
	if axe.damage != 2 {
		t.Errorf("axe damage = %d, want 2", axe.damage)
	}
}

func TestCopperHoneycombAppliesWax(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCopper(w)

	honeycomb := fakeItem{typeID: itemTypeIDsHoneycomb}
	if !c.OnInteract(honeycomb, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to apply wax")
	}
	if !c.IsWaxed() {
		t.Error("honeycomb should have waxed the block")
	}
}
