package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

var _ inventory.Inventory = (*CraftingTableInventory)(nil)

func newTestCraftingItem() item.Item {
	return item.NewApple(item.NewItemIdentifier(item.APPLE), "Apple")
}

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

func TestCraftingGridGetIngredientPanicsWhenEmpty(t *testing.T) {
	c := NewCraftingGrid(CraftingGridSizeSmall)
	defer func() {
		if recover() == nil {
			t.Error("expected GetIngredient to panic on an empty grid")
		}
	}()
	c.GetIngredient(0, 0)
}

func TestCraftingGridTracksBoundsOfOccupiedSlots(t *testing.T) {
	c := NewCraftingGrid(CraftingGridSizeBig)
	// place items at (1,0) and (2,1) in a 3-wide grid -> slots 1 and 5
	c.SetItem(1, newTestCraftingItem())
	c.SetItem(5, newTestCraftingItem())

	if c.GetRecipeWidth() != 2 {
		t.Errorf("GetRecipeWidth() = %d, want 2 (columns 1-2)", c.GetRecipeWidth())
	}
	if c.GetRecipeHeight() != 2 {
		t.Errorf("GetRecipeHeight() = %d, want 2 (rows 0-1)", c.GetRecipeHeight())
	}

	ingredient := c.GetIngredient(0, 0) // offset (0,0) from the top-left of the bounding box -> slot 1
	if ingredient.IsNull() {
		t.Error("expected GetIngredient(0,0) to find the item placed at the grid's top-left bound")
	}
}

func TestCraftingGridResetsBoundsWhenCleared(t *testing.T) {
	c := NewCraftingGrid(CraftingGridSizeSmall)
	c.SetItem(0, newTestCraftingItem())
	c.Clear(0)

	defer func() {
		if recover() == nil {
			t.Error("expected GetIngredient to panic once the grid is empty again")
		}
	}()
	c.GetIngredient(0, 0)
}
