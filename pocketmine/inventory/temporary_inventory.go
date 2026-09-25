package inventory

// TemporaryInventory is a port of pocketmine\inventory\TemporaryInventory: an inventory whose
// contents are given back to (or dropped by) the player when they close it, e.g. the cursor and
// the crafting grid.
type TemporaryInventory interface {
	Inventory
	// IsTemporaryInventory marks the type (PHP's `implements TemporaryInventory`).
	IsTemporaryInventory()
}
