package item

import (
	"slices"
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/utils"
)

func newEmptyStringToItemParser() *StringToItemParser {
	return &StringToItemParser{StringToTParser: utils.NewStringToTParser[Item](), reverseMap: map[int]map[string]bool{}}
}

// The three tests below are StringToItemParserTest.php.

func TestStringToItemParserOverrideRemovesOldAlias(t *testing.T) {
	p := newEmptyStringToItemParser()
	item1, item2 := VanillaItem("diamond"), VanillaItem("emerald")
	p.mustRegister("test_alias", func(string) Item { return item1 })
	if !slices.Contains(p.LookupAliases(item1), "test_alias") || slices.Contains(p.LookupAliases(item2), "test_alias") {
		t.Fatal("alias registered under the wrong item")
	}
	p.Override("test_alias", func(string) Item { return item2 })
	if slices.Contains(p.LookupAliases(item1), "test_alias") || !slices.Contains(p.LookupAliases(item2), "test_alias") {
		t.Fatal("override didn't move the alias")
	}
}

func TestStringToItemParserOverrideWithNewAlias(t *testing.T) {
	p := newEmptyStringToItemParser()
	it := VanillaItem("diamond")
	p.Override("new_alias", func(string) Item { return it })
	if !slices.Contains(p.LookupAliases(it), "new_alias") {
		t.Fatal("new alias missing from the reverse map")
	}
	if got, ok := p.Parse("new_alias"); !ok || got != it {
		t.Fatal("parse didn't return the overriding item")
	}
}

func TestStringToItemParserOverrideMultipleAliases(t *testing.T) {
	p := newEmptyStringToItemParser()
	item1, item2 := VanillaItem("diamond"), VanillaItem("emerald")
	for _, a := range []string{"alias1", "alias2", "alias3"} {
		p.mustRegister(a, func(string) Item { return item1 })
	}
	if n := len(p.LookupAliases(item1)); n != 3 {
		t.Fatalf("%d aliases, want 3", n)
	}
	p.Override("alias2", func(string) Item { return item2 })
	if got := p.LookupAliases(item1); !slices.Equal(got, []string{"alias1", "alias3"}) {
		t.Fatalf("item1 aliases = %v", got)
	}
	if got := p.LookupAliases(item2); !slices.Equal(got, []string{"alias2"}) {
		t.Fatalf("item2 aliases = %v", got)
	}
}

// TestStringToItemParserDefaults builds the whole built-in table (every block and item name and
// setter in it must exist) and checks some aliases.
func TestStringToItemParserDefaults(t *testing.T) {
	p := GetStringToItemParser()

	sword, ok := p.Parse("minecraft:Diamond Sword")
	if !ok || sword.GetStateId() != VanillaItem("diamond_sword").GetStateId() {
		t.Errorf("diamond sword: %v %v", sword, ok)
	}
	if _, ok := p.Parse("no_such_item"); ok {
		t.Error("unknown alias parsed")
	}

	wool, ok := p.Parse("red_wool")
	if !ok {
		t.Fatal("red_wool missing")
	}
	blk := wool.(*ItemBlock).Block
	if c := blk.(interface{ GetColor() blockutils.DyeColor }).GetColor(); c != blockutils.DyeColorRed {
		t.Errorf("red_wool colour = %v", c)
	}

	dye, ok := p.Parse("light_blue_dye")
	if !ok || dye.(*Dye).GetColor() != blockutils.DyeColorLightBlue {
		t.Errorf("light_blue_dye: %v", dye)
	}
	if _, ok := p.Parse("waxed_oxidized_cut_copper_stairs"); !ok {
		t.Error("waxed_oxidized_cut_copper_stairs missing")
	}
	if _, ok := p.Parse("strong_healing_splash_potion"); !ok {
		t.Error("strong_healing_splash_potion missing")
	}

	if !slices.Contains(p.LookupBlockAliases(block.VanillaBlock("stone")), "stone") {
		t.Error("stone isn't an alias of the stone block")
	}
}
