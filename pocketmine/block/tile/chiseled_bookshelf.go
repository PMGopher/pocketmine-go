package tile

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const chiseledBookshelfTagLastInteractedSlot = "LastInteractedSlot"

// ChiseledBookshelf is a port of pocketmine\block\tile\ChiseledBookshelf. Unlike every other
// container tile, it extends Tile directly (not Spawnable), and its items are saved as a
// positional list (empty slots included).
type ChiseledBookshelf struct {
	TileBase
	ContainerComponent

	lastInteractedSlot *blockutils.ChiseledBookshelfSlot
}

func NewChiseledBookshelf(world World, pos math.Vector3) *ChiseledBookshelf {
	c := &ChiseledBookshelf{}
	c.TileBase = NewTileBase(world, pos)
	c.Init(c)
	return c
}

func (c *ChiseledBookshelf) SaveID() string { return "ChiseledBookshelf" }

func (c *ChiseledBookshelf) GetLastInteractedSlot() (blockutils.ChiseledBookshelfSlot, bool) {
	if c.lastInteractedSlot == nil {
		return 0, false
	}
	return *c.lastInteractedSlot, true
}

func (c *ChiseledBookshelf) SetLastInteractedSlot(slot *blockutils.ChiseledBookshelfSlot) {
	c.lastInteractedSlot = slot
}

// GetInventory is a port of ChiseledBookshelf::getInventory (a 6-slot SimpleInventory).
func (c *ChiseledBookshelf) GetInventory() Inventory { return c.realInventory(c) }

// GetRealInventory is a port of ChiseledBookshelf::getRealInventory.
func (c *ChiseledBookshelf) GetRealInventory() Inventory { return c.realInventory(c) }

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (c *ChiseledBookshelf) OnBlockDestroyedHook() { c.dropContents(c) }

// loadItems is a port of ChiseledBookshelf::loadItems: list positions are slots, and entries with
// a count of 0 are empty slots.
func (c *ChiseledBookshelf) loadItems(tag *nbt.CompoundTag) {
	if list, ok, err := tag.GetListTag(ContainerTagItems); err == nil && ok && (list.Count() == 0 || list.GetTagType() == nbt.TagCompound) {
		if inv := c.realInventory(c); inv != nil && LoadInventoryItemsFunc != nil {
			var items []*nbt.CompoundTag
			for slot, t := range list.Values() {
				itemTag := t.(*nbt.CompoundTag)
				if itemTag.GetByteOr("Count", 0) == 0 {
					continue
				}
				withSlot := itemTag.Clone()
				withSlot.SetByte("Slot", nbt.ByteTag(slot))
				items = append(items, withSlot)
			}
			LoadInventoryItemsFunc(inv, items, fmt.Sprintf("ChiseledBookshelf (%v)", c.position.Vector3))
		}
	}
	if t, ok := tag.GetTag(ContainerTagLock); ok {
		if lock, ok := t.(nbt.StringTag); ok {
			c.Lock, c.HasLock = string(lock), true
		}
	}
}

// saveItems is a port of ChiseledBookshelf::saveItems.
func (c *ChiseledBookshelf) saveItems(tag *nbt.CompoundTag) {
	if inv := c.realInventory(c); inv != nil && SaveInventoryItemsFunc != nil {
		values := make([]nbt.Tag, blockutils.ChiseledBookshelfSlotCount)
		for _, itemTag := range SaveInventoryItemsFunc(inv) {
			slot := int(itemTag.GetByteOr("Slot", 0))
			if slot < 0 || slot >= len(values) {
				continue
			}
			itemTag = itemTag.Clone()
			itemTag.RemoveTag("Slot")
			values[slot] = itemTag
		}
		for slot, v := range values {
			if v == nil {
				values[slot] = nbt.NewCompoundTag().
					SetByte("Count", 0).
					SetShort("Damage", 0).
					SetString("Name", "").
					SetByte("WasPickedUp", 0)
			}
		}
		list, _ := nbt.NewListTag(values, nbt.TagCompound)
		tag.SetTag(ContainerTagItems, list)
	}
	if c.HasLock {
		tag.SetString(ContainerTagLock, nbt.StringTag(c.Lock))
	}
}

// ReadSaveData is a port of ChiseledBookshelf::readSaveData.
func (c *ChiseledBookshelf) ReadSaveData(tag *nbt.CompoundTag) error {
	c.loadItems(tag)
	raw := int(tag.GetIntOr(chiseledBookshelfTagLastInteractedSlot, 0))
	if raw != 0 {
		slot := blockutils.ChiseledBookshelfSlot(raw - 1)
		c.lastInteractedSlot = &slot
	}
	return nil
}

// WriteSaveData is a port of ChiseledBookshelf::writeSaveData.
func (c *ChiseledBookshelf) WriteSaveData(tag *nbt.CompoundTag) {
	c.saveItems(tag)
	value := 0
	if c.lastInteractedSlot != nil {
		value = int(*c.lastInteractedSlot) + 1
	}
	tag.SetInt(chiseledBookshelfTagLastInteractedSlot, nbt.IntTag(value))
}
