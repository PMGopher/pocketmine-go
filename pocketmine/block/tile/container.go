package tile

const (
	ContainerTagItems = "Items"
	ContainerTagLock  = "Lock"
)

// ContainerComponent is a port of the lock/canOpenWith part of pocketmine\block\tile\
// ContainerTrait.
//
// The rest of ContainerTrait isn't ported here: GetRealInventory (and the Container interface
// itself, which PHP declares as `getRealInventory(): Inventory`) needs the inventory package, but
// inventory imports item, and item imports block, and block imports tile - so tile importing
// inventory at all would cycle. This is the same category of constraint that keeps tile from
// importing block directly (see this package's doc comment); fixing it for real would mean either
// item no longer importing block (a decent-sized refactor across every TieredTool-family file) or
// restructuring which package block/inventory's container types live in. Until then, a concrete
// container tile's slot storage isn't wired up in this package - see e.g. Chest's doc comment.
//
// loadItems/saveItems (the item NBT round trip) also aren't ported regardless - see item.Item's
// doc comment on NbtSerialize/SafeNbtDeserialize needing the unported GlobalItemDataHandlers
// registry.
type ContainerComponent struct {
	Lock    string
	HasLock bool
}

// CanOpenWith is a port of ContainerTrait::canOpenWith.
func (c *ContainerComponent) CanOpenWith(key string) bool {
	return !c.HasLock || c.Lock == key
}
