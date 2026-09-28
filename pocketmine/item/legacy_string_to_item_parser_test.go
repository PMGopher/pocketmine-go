package item_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	_ "pocketmine-go/pocketmine/world/format/io"
)

// TestLegacyStringToItemParser is LegacyStringToItemParserTest.php.
func TestLegacyStringToItemParser(t *testing.T) {
	diorite, err := block.VanillaBlock("diorite").(interface{ AsItem() (block.Item, error) }).AsItem()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		input    string
		expected item.Item
	}{
		{"dye:4", item.VanillaItem("lapis_lazuli")},
		{"351", item.VanillaItem("ink_sac")},
		{"351:4", item.VanillaItem("lapis_lazuli")},
		{"stone:3", diorite.(item.Item)},
		{"minecraft:string", item.VanillaItem("string")},
		{"diamond_pickaxe", item.VanillaItem("diamond_pickaxe")},
	} {
		got, err := item.GetLegacyStringToItemParser().Parse(c.input)
		if err != nil {
			t.Errorf("%s: %v", c.input, err)
			continue
		}
		if !got.Equals(c.expected, true) {
			t.Errorf("%s = %v, want %v", c.input, got, c.expected)
		}
	}
	if _, err := item.GetLegacyStringToItemParser().Parse("stone:abc"); err == nil {
		t.Error("non-numeric meta accepted")
	}
}
