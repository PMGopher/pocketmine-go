package player

import (
	"pocketmine-go/pocketmine/block"
	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/inventory"
)

// init wires block.OpenWindowFunc: a block's onInteract doing
// $player->setCurrentWindow(new XInventory($this->position)).
func init() {
	// $player->setCurrentWindow($tile->getInventory()) from a container block's onInteract.
	block.OpenTileWindowFunc = func(who block.Player, inv tile.Inventory) bool {
		p, ok := who.(*Player)
		if !ok {
			return false
		}
		window, ok := inv.(inventory.Inventory)
		if !ok {
			return false
		}
		return p.SetCurrentWindow(window)
	}
	block.OpenWindowFunc = func(who block.Player, windowType block.WindowType, pos block.Position) bool {
		p, ok := who.(*Player)
		if !ok {
			return false
		}
		var inv inventory.Inventory
		switch windowType {
		case block.WindowAnvil:
			inv = blockinventory.NewAnvilInventory(pos)
		case block.WindowCartographyTable:
			inv = blockinventory.NewCartographyTableInventory(pos)
		case block.WindowCraftingTable:
			inv = blockinventory.NewCraftingTableInventory(pos)
		case block.WindowEnchantingTable:
			inv = blockinventory.NewEnchantInventory(pos)
		case block.WindowEnderChest:
			inv = blockinventory.NewEnderChestInventory(pos, p.GetEnderInventory())
		case block.WindowLoom:
			inv = blockinventory.NewLoomInventory(pos)
		case block.WindowSmithingTable:
			inv = blockinventory.NewSmithingTableInventory(pos)
		case block.WindowStonecutter:
			inv = blockinventory.NewStonecutterInventory(pos)
		default:
			return false
		}
		return p.SetCurrentWindow(inv)
	}
}
