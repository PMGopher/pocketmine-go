package inventory

// InventoryHolder is a port of pocketmine\inventory\InventoryHolder.
type InventoryHolder interface {
	GetInventory() Inventory
}
