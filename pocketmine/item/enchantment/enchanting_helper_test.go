package enchantment

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

type fakeEnchantableItem struct {
	tags           []string
	enchanted      bool
	enchantability int
}

func (f fakeEnchantableItem) GetEnchantmentTags() []string { return f.tags }
func (f fakeEnchantableItem) HasEnchantments() bool        { return f.enchanted }
func (f fakeEnchantableItem) IsNull() bool                 { return false }
func (f fakeEnchantableItem) GetEnchantability() int       { return f.enchantability }

// bookshelfRing returns surroundings with air around the table and bookshelves in the first n
// ring positions (at y=0) in countBookshelves' iteration order.
func bookshelfRing(n int) TableSurroundings {
	shelves := map[[2]int]bool{}
	for x := -2; x <= 2 && len(shelves) < n; x++ {
		for z := -2; z <= 2 && len(shelves) < n; z++ {
			if abs(x) == 2 || abs(z) == 2 {
				shelves[[2]int{x, z}] = true
			}
		}
	}
	return func(dx, dy, dz int) (bool, bool) {
		if dy == 0 && shelves[[2]int{dx, dz}] {
			return false, true
		}
		return true, false
	}
}

func TestCountBookshelves(t *testing.T) {
	if got := countBookshelves(bookshelfRing(7)); got != 7 {
		t.Errorf("7 bookshelves: got %d", got)
	}
	// Two layers on every one of the 16 ring positions is capped at 15.
	full := func(dx, dy, dz int) (bool, bool) {
		ring := abs(dx) == 2 || abs(dz) == 2
		return !ring, ring
	}
	if got := countBookshelves(full); got != maxBookshelfCount {
		t.Errorf("full ring: got %d, want %d", got, maxBookshelfCount)
	}
	// Nothing counts when the space between the table and the shelves is blocked.
	blocked := func(dx, dy, dz int) (bool, bool) { return false, abs(dx) == 2 || abs(dz) == 2 }
	if got := countBookshelves(blocked); got != 0 {
		t.Errorf("blocked: got %d, want 0", got)
	}
}

// TestGenerateOptionsMatchesPHP compares GenerateOptions for a sword (enchantability 14) with
// the output of EnchantingHelper::generateOptions' algorithm run in PHP 8.4 with pocketmine's
// Random.
func TestGenerateOptionsMatchesPHP(t *testing.T) {
	golden := []struct {
		seed, bookshelves, level int
		name, enchantments       string
	}{
		{0, 0, 1, "nsdsvatkgcouei", "sharpness:1"},
		{0, 0, 3, "aqpzgkr", "knockback:1,sharpness:1"},
		{0, 0, 3, "huizirvmiazqgwe", "sharpness:1"},
		{0, 7, 4, "nsdsvatkgcouei", "knockback:1"},
		{0, 7, 9, "aqpzgkr", "knockback:1,sharpness:2"},
		{0, 7, 14, "huizirvmiazqgwe", "sharpness:2"},
		{0, 15, 5, "nsdsvatkgcouei", "knockback:1"},
		{0, 15, 11, "aqpzgkr", "knockback:1,sharpness:2"},
		{0, 15, 30, "izirv", "sharpness:4,unbreaking:3"},
		{12345, 0, 2, "khsdmsyb", "knockback:1"},
		{12345, 0, 6, "shtcdrbqzvlscgu", "sharpness:1"},
		{12345, 0, 8, "biodbbspakovdz", "unbreaking:2"},
		{12345, 7, 6, "khsdmsyb", "sharpness:1"},
		{12345, 7, 13, "shtcdrbqzvlscgu", "sharpness:2"},
		{12345, 7, 18, "biodbbspakovdz", "unbreaking:3"},
		{12345, 15, 7, "khsdmsyb", "sharpness:2"},
		{12345, 15, 15, "shtcdrbqzvlscgu", "sharpness:2"},
		{12345, 15, 30, "biodbbspakovdz", "unbreaking:3"},
		{-987654321, 0, 1, "csutoxyzibz", "sharpness:1"},
		{-987654321, 0, 1, "lqjbe", "knockback:1,sharpness:1"},
		{-987654321, 0, 1, "wwbnsj", "sharpness:1"},
		{-987654321, 7, 2, "csutoxyzibz", "sharpness:1"},
		{-987654321, 7, 5, "lqjbe", "unbreaking:1,sharpness:1"},
		{-987654321, 7, 14, "sjdzusuyrzwlaso", "sharpness:2,unbreaking:2,knockback:1"},
		{-987654321, 15, 3, "csutoxyzibz", "sharpness:1"},
		{-987654321, 15, 7, "lqjbe", "unbreaking:2,sharpness:2"},
		{-987654321, 15, 30, "sjdzusuyrzwlaso", "sharpness:4,unbreaking:3,knockback:2"},
		{2147483647, 0, 1, "zpiauo", "unbreaking:1"},
		{2147483647, 0, 2, "aktltgulqrrbk", "knockback:1"},
		{2147483647, 0, 2, "pgyssqdjrqqcokb", "sharpness:1"},
		{2147483647, 7, 3, "iauoqt", "unbreaking:1,knockback:1"},
		{2147483647, 7, 7, "tltgulq", "sharpness:2"},
		{2147483647, 7, 14, "xdpgyssqdj", "knockback:1,sharpness:2"},
		{2147483647, 15, 7, "iauoqt", "sharpness:2,unbreaking:1"},
		{2147483647, 15, 15, "tltgulq", "sharpness:2"},
		{2147483647, 15, 30, "xdpgyssqdj", "knockback:2,sharpness:4"},
	}
	sword := fakeEnchantableItem{tags: []string{TagSword}, enchantability: 14}
	for i := 0; i < len(golden); i += 3 {
		g := golden[i]
		options := GenerateOptions(bookshelfRing(g.bookshelves), sword, g.seed)
		if len(options) != 3 {
			t.Fatalf("seed %d: %d options", g.seed, len(options))
		}
		for j, option := range options {
			want := golden[i+j]
			var enchantments []string
			for _, e := range option.GetEnchantments() {
				enchantments = append(enchantments, fmt.Sprintf("%s:%d", swordEnchantmentNames[e.GetType()], e.GetLevel()))
			}
			got := strings.Join(enchantments, ",")
			if option.GetRequiredXpLevel() != want.level || option.GetDisplayName() != want.name || got != want.enchantments {
				t.Errorf("seed %d, %d bookshelves, option %d: got {%d %q %q}, want {%d %q %q}", want.seed, want.bookshelves, j,
					option.GetRequiredXpLevel(), option.GetDisplayName(), got, want.level, want.name, want.enchantments)
			}
		}
	}
}

var swordEnchantmentNames = map[Enchantment]string{
	VanillaSharpness():  "sharpness",
	VanillaKnockback():  "knockback",
	VanillaFireAspect(): "fire_aspect",
	VanillaUnbreaking(): "unbreaking",
}

func TestGenerateOptionsSkipsEnchantedItems(t *testing.T) {
	if options := GenerateOptions(bookshelfRing(0), fakeEnchantableItem{tags: []string{TagSword}, enchanted: true, enchantability: 14}, 1); len(options) != 0 {
		t.Errorf("an enchanted item got %d options", len(options))
	}
}

func TestItemEnchantmentTagRegistryIntersection(t *testing.T) {
	r := GetItemEnchantmentTagRegistry()
	cases := []struct {
		first, second []string
		want          bool
	}{
		{[]string{TagArmor}, []string{TagHelmet}, true},
		{[]string{TagWeapons}, []string{TagAxe}, true},
		{[]string{TagBlockTools}, []string{TagSword}, false},
		{[]string{TagAll}, []string{TagBow}, true},
		{[]string{TagSword}, []string{TagBow}, false},
		{nil, []string{TagBow}, false},
	}
	for _, c := range cases {
		if got := r.IsTagArrayIntersection(c.first, c.second); got != c.want {
			t.Errorf("IsTagArrayIntersection(%v, %v) = %v, want %v", c.first, c.second, got, c.want)
		}
	}
}

func TestAvailableEnchantmentRegistryPrimaryForSword(t *testing.T) {
	got := GetAvailableEnchantmentRegistry().GetPrimaryEnchantmentsForItem(fakeEnchantableItem{tags: []string{TagSword}})
	want := []Enchantment{VanillaSharpness(), VanillaKnockback(), VanillaFireAspect(), VanillaUnbreaking()}
	if !slices.Equal(got, want) {
		t.Errorf("primary sword enchantments: got %d, want %d (in registration order)", len(got), len(want))
	}
	all := GetAvailableEnchantmentRegistry().GetAllEnchantmentsForItem(fakeEnchantableItem{tags: []string{TagSword}})
	if !slices.Contains(all, VanillaMending()) || !slices.Contains(all, VanillaVanishing()) {
		t.Error("mending and curse of vanishing should be available for a sword as secondary enchantments")
	}
}
