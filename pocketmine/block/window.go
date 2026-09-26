package block

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

// openWindow calls OpenWindowFunc if it is set.
func openWindow(player Player, windowType WindowType, pos Position) bool {
	if OpenWindowFunc == nil || player == nil {
		return false
	}
	return OpenWindowFunc(player, windowType, pos)
}
