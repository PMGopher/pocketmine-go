package blockinventory

import (
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

const (
	CraftingGridSizeSmall = 2
	CraftingGridSizeBig   = 3
)

// CraftingGrid is a port of pocketmine\crafting\CraftingGrid, placed here rather than a separate
// crafting package since CraftingTableInventory is its only consumer so far and nothing else
// needs it yet - actual recipe matching (CraftingManager) isn't ported, but the grid's own
// bookkeeping (tracking the smallest rectangle of non-empty slots, so a recipe matcher would know
// where to look) is fully real and doesn't depend on recipe matching at all.
type CraftingGrid struct {
	*inventory.SimpleInventory

	gridWidth int

	hasBounds    bool
	startX, xLen int
	startY, yLen int
}

func NewCraftingGrid(gridWidth int) *CraftingGrid {
	c := &CraftingGrid{SimpleInventory: inventory.NewSimpleInventory(gridWidth * gridWidth), gridWidth: gridWidth}
	// Re-run Init so BaseInventory's internal self-dispatch (used by Clear/AddItem/RemoveItem/
	// Swap to reach the *current* SetItem override) targets CraftingGrid itself rather than the
	// bare *SimpleInventory NewSimpleInventory pointed it at - otherwise SetItem's
	// seekRecipeBounds() call would silently get skipped by every one of those.
	c.Init(c)
	return c
}

func (c *CraftingGrid) GetGridWidth() int { return c.gridWidth }

// SetItem shadows the promoted inventory.BaseInventory.SetItem so seekRecipeBounds runs on every
// change, the same "define here rather than rely on promotion" reasoning used throughout this
// port when a Go embedding doesn't get PHP's automatic $this-> dispatch (see
// AnimatedInventory.OnOpen's doc comment for another instance of the same fix).
func (c *CraftingGrid) SetItem(index int, it item.Item) {
	c.SimpleInventory.SetItem(index, it)
	c.seekRecipeBounds()
}

func (c *CraftingGrid) seekRecipeBounds() {
	minX, maxX := c.gridWidth, 0
	minY, maxY := c.gridWidth, 0
	empty := true

	for y := 0; y < c.gridWidth; y++ {
		for x := 0; x < c.gridWidth; x++ {
			if !c.IsSlotEmpty(y*c.gridWidth + x) {
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
				empty = false
			}
		}
	}

	if !empty {
		c.startX, c.xLen = minX, maxX-minX+1
		c.startY, c.yLen = minY, maxY-minY+1
		c.hasBounds = true
	} else {
		c.startX, c.xLen, c.startY, c.yLen, c.hasBounds = 0, 0, 0, 0, false
	}
}

// GetIngredient panics if the grid is empty, mirroring the PHP original's LogicException (a
// programmer error at the call site).
func (c *CraftingGrid) GetIngredient(x, y int) item.Item {
	if !c.hasBounds {
		panic("No ingredients found in grid")
	}
	return c.GetItem((y+c.startY)*c.gridWidth + (x + c.startX))
}

func (c *CraftingGrid) GetRecipeWidth() int { return c.xLen }

func (c *CraftingGrid) GetRecipeHeight() int { return c.yLen }
