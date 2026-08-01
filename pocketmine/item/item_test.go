package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/nbt"
)

// Compile-time proof that concrete item types satisfy both this package's Item interface and
// block's forward-compatible local interfaces (block.Item, block.Dye) via structural typing -
// this is the whole point of porting Item as its own package rather than inside block/.
var (
	_ Item       = (*Dye)(nil)
	_ Item       = (*Fertilizer)(nil)
	_ block.Item = (*Dye)(nil)
	_ block.Item = (*Fertilizer)(nil)
	_ block.Dye  = (*Dye)(nil)
)

func newTestDye() *Dye {
	return NewDye(NewItemIdentifier(DYE), "Dye")
}

func TestDyeGetColorDefaultsToBlack(t *testing.T) {
	d := newTestDye()
	if d.GetColor() != blockutils.DyeColorBlack {
		t.Errorf("GetColor() = %v, want Black", d.GetColor())
	}
}

func TestDyeCloneIsIndependent(t *testing.T) {
	d := newTestDye()
	d.SetColor(blockutils.DyeColorRed)

	clone := d.Clone().(*Dye)
	clone.SetColor(blockutils.DyeColorBlue)

	if d.GetColor() != blockutils.DyeColorRed {
		t.Errorf("original GetColor() = %v, want Red (clone mutation leaked)", d.GetColor())
	}
	if clone.GetColor() != blockutils.DyeColorBlue {
		t.Errorf("clone GetColor() = %v, want Blue", clone.GetColor())
	}
}

func TestDyeStateIdChangesWithColor(t *testing.T) {
	d := newTestDye()
	d.SetColor(blockutils.DyeColorRed)
	red := d.GetStateId()

	d.SetColor(blockutils.DyeColorBlue)
	blue := d.GetStateId()

	if red == blue {
		t.Error("expected different colors to produce different state IDs")
	}
}

func TestItemPopReducesCountByOne(t *testing.T) {
	d := newTestDye()
	d.SetCount(5)
	d.Pop()
	if d.GetCount() != 4 {
		t.Errorf("GetCount() = %d, want 4", d.GetCount())
	}
}

func TestItemPopCountSplitsStack(t *testing.T) {
	d := newTestDye()
	d.SetCount(5)

	split := d.PopCount(2)
	if d.GetCount() != 3 {
		t.Errorf("original GetCount() = %d, want 3", d.GetCount())
	}
	if split.GetCount() != 2 {
		t.Errorf("split GetCount() = %d, want 2", split.GetCount())
	}
}

func TestItemPopCountPanicsWhenExceedingStack(t *testing.T) {
	d := newTestDye()
	d.SetCount(1)

	defer func() {
		if recover() == nil {
			t.Error("expected PopCount to panic when count exceeds the stack")
		}
	}()
	d.PopCount(2)
}

func TestItemIsNull(t *testing.T) {
	d := newTestDye()
	if d.IsNull() {
		t.Error("expected a fresh item with count 1 not to be null")
	}
	d.SetCount(0)
	if !d.IsNull() {
		t.Error("expected an item with count 0 to be null")
	}
}

func TestItemCustomNameRoundTripsThroughNBT(t *testing.T) {
	d := newTestDye()
	d.SetCustomName("My Dye")

	tag := d.GetNamedTag()

	decoded := newTestDye()
	decoded.SetNamedTag(tag)

	if !decoded.HasCustomName() || decoded.GetCustomName() != "My Dye" {
		t.Errorf("GetCustomName() = %q, want %q", decoded.GetCustomName(), "My Dye")
	}
	if decoded.GetName() != "My Dye" {
		t.Errorf("GetName() = %q, want the custom name", decoded.GetName())
	}
}

func TestItemLoreRoundTripsThroughNBT(t *testing.T) {
	d := newTestDye()
	d.SetLore([]string{"line one", "line two"})

	decoded := newTestDye()
	decoded.SetNamedTag(d.GetNamedTag())

	lore := decoded.GetLore()
	if len(lore) != 2 || lore[0] != "line one" || lore[1] != "line two" {
		t.Errorf("GetLore() = %v, want [line one, line two]", lore)
	}
}

func TestItemCanPlaceOnRoundTripsThroughNBT(t *testing.T) {
	d := newTestDye()
	d.SetCanPlaceOn([]string{"stone", "dirt"})

	decoded := newTestDye()
	decoded.SetNamedTag(d.GetNamedTag())

	places := decoded.GetCanPlaceOn()
	if len(places) != 2 || places["stone"] != "stone" || places["dirt"] != "dirt" {
		t.Errorf("GetCanPlaceOn() = %v, want stone and dirt", places)
	}
}

func TestItemKeepOnDeathRoundTripsThroughNBT(t *testing.T) {
	d := newTestDye()
	d.SetKeepOnDeath(true)

	decoded := newTestDye()
	decoded.SetNamedTag(d.GetNamedTag())

	if !decoded.KeepOnDeath() {
		t.Error("expected KeepOnDeath() to round trip as true")
	}
}

func TestItemCustomBlockDataRoundTripsThroughNBT(t *testing.T) {
	d := newTestDye()
	blockData := nbt.NewCompoundTag()
	blockData.SetString("Foo", "Bar")
	d.SetCustomBlockData(blockData)

	decoded := newTestDye()
	decoded.SetNamedTag(d.GetNamedTag())

	if !decoded.HasCustomBlockData() {
		t.Fatal("expected HasCustomBlockData() to be true after round trip")
	}
	if s, err := decoded.GetCustomBlockData().GetString("Foo"); err != nil || string(s) != "Bar" {
		t.Errorf("GetCustomBlockData()[Foo] = %q (err %v), want %q", s, err, "Bar")
	}
}

func TestItemClearNamedTagResetsState(t *testing.T) {
	d := newTestDye()
	d.SetCustomName("Named")
	d.SetLore([]string{"lore"})
	d.GetNamedTag() // force a serialize so the tag reflects the above

	d.ClearNamedTag()

	if d.HasCustomName() || len(d.GetLore()) != 0 {
		t.Error("expected ClearNamedTag to reset custom name and lore")
	}
	if d.HasNamedTag() {
		t.Error("expected HasNamedTag to be false after ClearNamedTag")
	}
}

func TestItemEqualsComparesStateAndNbt(t *testing.T) {
	a := newTestDye()
	a.SetColor(blockutils.DyeColorRed)
	b := newTestDye()
	b.SetColor(blockutils.DyeColorRed)

	if !a.Equals(b, true) {
		t.Error("expected two same-colored dyes with no NBT to be equal")
	}

	b.SetColor(blockutils.DyeColorBlue)
	if a.Equals(b, true) {
		t.Error("expected differently-colored dyes not to be equal")
	}
}

func TestItemEqualsExactRequiresSameCount(t *testing.T) {
	a := newTestDye()
	a.SetCount(2)
	b := newTestDye()
	b.SetCount(3)

	if a.EqualsExact(b) {
		t.Error("expected EqualsExact to require matching count")
	}
	b.SetCount(2)
	if !a.EqualsExact(b) {
		t.Error("expected EqualsExact to pass once count matches")
	}
}

func TestFertilizerSatisfiesItem(t *testing.T) {
	f := NewFertilizer(NewItemIdentifier(BONE_MEAL), "Bone Meal")
	if f.GetTypeId() != BONE_MEAL {
		t.Errorf("GetTypeId() = %d, want %d", f.GetTypeId(), BONE_MEAL)
	}
	if f.GetMaxStackSize() != 64 {
		t.Errorf("GetMaxStackSize() = %d, want 64", f.GetMaxStackSize())
	}
}
