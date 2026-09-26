package blockinventory

import "pocketmine-go/pocketmine/inventory"

// The block inventories PHP declares `implements TemporaryInventory`: their contents are given
// back to the player (or dropped) when the window closes (Player::doCloseInventory).

func (a *AnvilInventory) IsTemporaryInventory()            {}
func (c *CartographyTableInventory) IsTemporaryInventory() {}
func (c *CraftingTableInventory) IsTemporaryInventory()    {}
func (e *EnchantInventory) IsTemporaryInventory()          {}
func (l *LoomInventory) IsTemporaryInventory()             {}
func (s *SmithingTableInventory) IsTemporaryInventory()    {}
func (s *StonecutterInventory) IsTemporaryInventory()      {}

var (
	_ inventory.TemporaryInventory = (*AnvilInventory)(nil)
	_ inventory.TemporaryInventory = (*CartographyTableInventory)(nil)
	_ inventory.TemporaryInventory = (*CraftingTableInventory)(nil)
	_ inventory.TemporaryInventory = (*EnchantInventory)(nil)
	_ inventory.TemporaryInventory = (*LoomInventory)(nil)
	_ inventory.TemporaryInventory = (*SmithingTableInventory)(nil)
	_ inventory.TemporaryInventory = (*StonecutterInventory)(nil)
)
