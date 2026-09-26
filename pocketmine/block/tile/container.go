package tile

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	ContainerTagItems = "Items"
	ContainerTagLock  = "Lock"
)

// Inventory is a container tile's inventory: an inventory.Inventory. This package can't import the
// inventory package (it imports item, which imports block, which imports this package), so the
// inventory is opaque here and handled through the container hooks below, which block/inventory
// sets in its init().
type Inventory any

// Container hooks, set by block/inventory.
var (
	// NewInventoryFunc is `new XInventory($this->position)` in a container tile's constructor: the
	// tile's own inventory type (ChestInventory, FurnaceInventory, ...).
	NewInventoryFunc func(t Tile) Inventory
	// LoadInventoryItemsFunc is ContainerTrait::loadItems' contents part: setContents with the
	// deserialized items, without firing the inventory's listeners. errorLogContext describes the
	// container for Item::safeNbtDeserialize's error messages.
	LoadInventoryItemsFunc func(inv Inventory, items []*nbt.CompoundTag, errorLogContext string)
	// SaveInventoryItemsFunc is ContainerTrait::saveItems' contents part: every item, serialized
	// with its slot.
	SaveInventoryItemsFunc func(inv Inventory) []*nbt.CompoundTag
	// DropInventoryContentsFunc is ContainerTrait::onBlockDestroyedHook: drops every item at pos
	// and clears the inventory.
	DropInventoryContentsFunc func(inv Inventory, world World, pos math.Vector3)
	// RemoveAllViewersFunc is Inventory::removeAllViewers.
	RemoveAllViewersFunc func(inv Inventory)
	// InventoryIsSlotEmptyFunc is Inventory::isSlotEmpty.
	InventoryIsSlotEmptyFunc func(inv Inventory, slot int) bool
)

// Container is a port of pocketmine\block\tile\Container: a tile with an inventory.
type Container interface {
	Tile
	GetRealInventory() Inventory
	CanOpenWith(key string) bool
}

// ContainerComponent is a port of pocketmine\block\tile\ContainerTrait: the inventory (created
// lazily through NewInventoryFunc, see Inventory) and lock of a container tile.
type ContainerComponent struct {
	Lock    string
	HasLock bool

	inventory Inventory
}

// realInventory is getRealInventory(): owner's inventory, created on first use.
func (c *ContainerComponent) realInventory(owner Tile) Inventory {
	if c.inventory == nil && NewInventoryFunc != nil {
		c.inventory = NewInventoryFunc(owner)
	}
	return c.inventory
}

// loadItems is a port of ContainerTrait::loadItems.
func (c *ContainerComponent) loadItems(owner Tile, tag *nbt.CompoundTag) {
	c.loadItemList(owner, tag, fmt.Sprintf("Container (%v)", owner.GetPosition().Vector3))
	if t, ok := tag.GetTag(ContainerTagLock); ok {
		if lock, ok := t.(nbt.StringTag); ok {
			c.Lock, c.HasLock = string(lock), true
		}
	}
}

// loadItemList reads the Items list into the inventory (a list of another type is ignored, which
// preserves PHP's old behaviour of not throwing on wrong types).
func (c *ContainerComponent) loadItemList(owner Tile, tag *nbt.CompoundTag, errorLogContext string) {
	list, ok, err := tag.GetListTag(ContainerTagItems)
	if err != nil || !ok || (list.Count() > 0 && list.GetTagType() != nbt.TagCompound) {
		return
	}
	inv := c.realInventory(owner)
	if inv == nil || LoadInventoryItemsFunc == nil {
		return
	}
	items := make([]*nbt.CompoundTag, 0, list.Count())
	for _, t := range list.Values() {
		items = append(items, t.(*nbt.CompoundTag))
	}
	LoadInventoryItemsFunc(inv, items, errorLogContext)
}

// saveItems is a port of ContainerTrait::saveItems.
func (c *ContainerComponent) saveItems(owner Tile, tag *nbt.CompoundTag) {
	if inv := c.realInventory(owner); inv != nil && SaveInventoryItemsFunc != nil {
		items := SaveInventoryItemsFunc(inv)
		values := make([]nbt.Tag, len(items))
		for i, it := range items {
			values[i] = it
		}
		list, _ := nbt.NewListTag(values, nbt.TagCompound)
		tag.SetTag(ContainerTagItems, list)
	}
	if c.HasLock {
		tag.SetString(ContainerTagLock, nbt.StringTag(c.Lock))
	}
}

// CanOpenWith is a port of ContainerTrait::canOpenWith.
func (c *ContainerComponent) CanOpenWith(key string) bool {
	return !c.HasLock || c.Lock == key
}

// dropContents is a port of ContainerTrait::onBlockDestroyedHook.
func (c *ContainerComponent) dropContents(owner Tile) {
	inv := c.realInventory(owner)
	pos := owner.GetPosition()
	world, ok := pos.GetWorld()
	if inv == nil || !ok || DropInventoryContentsFunc == nil {
		return
	}
	DropInventoryContentsFunc(inv, world, pos.Add(0.5, 0.5, 0.5))
}

// removeAllViewers is the `$this->inventory->removeAllViewers()` of the container tiles' close().
func (c *ContainerComponent) removeAllViewers() {
	if c.inventory != nil && RemoveAllViewersFunc != nil {
		RemoveAllViewersFunc(c.inventory)
	}
}
