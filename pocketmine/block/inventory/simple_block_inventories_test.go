package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

var (
	_ inventory.Inventory = (*LoomInventory)(nil)
	_ inventory.Inventory = (*CartographyTableInventory)(nil)
	_ inventory.Inventory = (*SmithingTableInventory)(nil)
	_ inventory.Inventory = (*StonecutterInventory)(nil)
	_ inventory.Inventory = (*EnchantInventory)(nil)
	_ BlockInventory      = (*LoomInventory)(nil)
	_ BlockInventory      = (*CartographyTableInventory)(nil)
	_ BlockInventory      = (*SmithingTableInventory)(nil)
	_ BlockInventory      = (*StonecutterInventory)(nil)
	_ BlockInventory      = (*EnchantInventory)(nil)
)

func TestSimpleBlockInventorySizes(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(0, 0, 0, w)

	cases := []struct {
		name string
		inv  inventory.Inventory
		want int
	}{
		{"Loom", NewLoomInventory(holder), 3},
		{"CartographyTable", NewCartographyTableInventory(holder), 2},
		{"SmithingTable", NewSmithingTableInventory(holder), 3},
		{"Stonecutter", NewStonecutterInventory(holder), 1},
		{"Enchant", NewEnchantInventory(holder), 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.inv.GetSize(); got != c.want {
				t.Errorf("GetSize() = %d, want %d", got, c.want)
			}
		})
	}
}

func TestSimpleBlockInventoriesShareHolder(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(3, 4, 5, w)
	loom := NewLoomInventory(holder)
	if loom.GetHolder() != holder {
		t.Errorf("GetHolder() = %v, want %v", loom.GetHolder(), holder)
	}
}

func TestEnchantInventoryInputAndLapisSlots(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(0, 0, 0, w)
	e := NewEnchantInventory(holder)

	book := item.NewApple(item.NewItemIdentifier(item.APPLE), "Apple")
	book.SetCount(1)
	e.SetItem(EnchantSlotInput, book)

	if e.GetInput().GetCount() != 1 {
		t.Errorf("GetInput().GetCount() = %d, want 1", e.GetInput().GetCount())
	}
	if !e.GetLapis().IsNull() {
		t.Error("expected the lapis slot to start empty")
	}
}

func TestLoomInventorySlotConstants(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(0, 0, 0, w)
	l := NewLoomInventory(holder)

	item1 := item.NewApple(item.NewItemIdentifier(item.APPLE), "Apple")
	l.SetItem(LoomSlotBanner, item1)
	if l.GetItem(0).IsNull() {
		t.Error("expected LoomSlotBanner (0) to hold the item that was set")
	}
	if LoomSlotDye != 1 || LoomSlotPattern != 2 {
		t.Errorf("LoomSlotDye=%d LoomSlotPattern=%d, want 1/2", LoomSlotDye, LoomSlotPattern)
	}
}
