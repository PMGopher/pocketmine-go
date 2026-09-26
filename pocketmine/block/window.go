package block

import "pocketmine-go/pocketmine/block/tile"

// WindowType names the block inventory a block opens for a player.
type WindowType int

const (
	WindowAnvil WindowType = iota
	WindowCartographyTable
	WindowCraftingTable
	WindowEnchantingTable
	WindowEnderChest
	WindowLoom
	WindowSmithingTable
	WindowStonecutter
)

// OpenWindowFunc is $player->setCurrentWindow(new XInventory($this->position)) from a block's
// onInteract: this package can't import block/inventory (which imports it) nor player, so the
// player package sets it in init() to build the block inventory and open it. It returns whether
// the window was opened. It is nil in tests that don't import player, and then nothing opens.
var OpenWindowFunc func(player Player, windowType WindowType, pos Position) bool

// OpenTileWindowFunc is $player->setCurrentWindow($tile->getInventory()) for a container tile's
// inventory (an inventory.Inventory, opaque here: see tile.Inventory). The player package sets it.
var OpenTileWindowFunc func(player Player, inv tile.Inventory) bool

// openTileWindow calls OpenTileWindowFunc if it is set.
func openTileWindow(player Player, inv tile.Inventory) bool {
	if OpenTileWindowFunc == nil || player == nil || inv == nil {
		return false
	}
	return OpenTileWindowFunc(player, inv)
}

// openWindow calls OpenWindowFunc if it is set.
func openWindow(player Player, windowType WindowType, pos Position) bool {
	if OpenWindowFunc == nil || player == nil {
		return false
	}
	return OpenWindowFunc(player, windowType, pos)
}
