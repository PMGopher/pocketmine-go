package blockinventory

import (
	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/inventory"
)

// PlayerCraftingInventory is a port of pocketmine\inventory\PlayerCraftingInventory: the 2x2
// crafting grid in the player's own inventory. It lives next to CraftingGrid (which it extends)
// because the inventory package can't import this one.
type PlayerCraftingInventory struct {
	*crafting.CraftingGrid

	holder inventory.Player
}

func NewPlayerCraftingInventory(holder inventory.Player) *PlayerCraftingInventory {
	p := &PlayerCraftingInventory{CraftingGrid: crafting.NewCraftingGrid(crafting.CraftingGridSizeSmall), holder: holder}
	// Viewers and listeners must see this inventory, not the embedded grid (see NewCraftingGrid).
	p.Init(p)
	return p
}

func (p *PlayerCraftingInventory) GetHolder() inventory.Player { return p.holder }

func (p *PlayerCraftingInventory) IsTemporaryInventory() {}

var _ inventory.TemporaryInventory = (*PlayerCraftingInventory)(nil)
