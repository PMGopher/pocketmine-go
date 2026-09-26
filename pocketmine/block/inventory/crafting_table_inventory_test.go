package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
)

var _ inventory.Inventory = (*CraftingTableInventory)(nil)

func TestCraftingTableInventorySize(t *testing.T) {
	w := &fakeWorld{}
	c := NewCraftingTableInventory(block.NewPosition(0, 0, 0, w))
	if c.GetSize() != 9 {
		t.Errorf("GetSize() = %d, want 9 (3x3)", c.GetSize())
	}
	if c.GetGridWidth() != 3 {
		t.Errorf("GetGridWidth() = %d, want 3", c.GetGridWidth())
	}
}
