package tile

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const chiseledBookshelfTagLastInteractedSlot = "LastInteractedSlot"

// ChiseledBookshelf is a port of pocketmine\block\tile\ChiseledBookshelf, minus its inventory/
// Container half - see ContainerComponent's doc comment for why the inventory package can't be
// imported here (loadItems/saveItems also need Item::safeNbtDeserialize/nbtSerialize, not ported
// either - same gap as every other container tile's item NBT). LastInteractedSlot is fully real.
//
// Unlike every other container tile in this port, the PHP original extends Tile directly (not
// Spawnable) - so this embeds TileBase, same as Note/Comparator/DaylightSensor.
type ChiseledBookshelf struct {
	TileBase

	lastInteractedSlot *blockutils.ChiseledBookshelfSlot
}

func NewChiseledBookshelf(world World, pos math.Vector3) *ChiseledBookshelf {
	c := &ChiseledBookshelf{}
	c.TileBase = NewTileBase(world, pos)
	c.Init(c)
	return c
}

func (c *ChiseledBookshelf) SaveID() string { return "Chiseled Bookshelf" }

func (c *ChiseledBookshelf) GetLastInteractedSlot() (blockutils.ChiseledBookshelfSlot, bool) {
	if c.lastInteractedSlot == nil {
		return 0, false
	}
	return *c.lastInteractedSlot, true
}

func (c *ChiseledBookshelf) SetLastInteractedSlot(slot *blockutils.ChiseledBookshelfSlot) {
	c.lastInteractedSlot = slot
}

// ReadSaveData is a port of ChiseledBookshelf::readSaveData, minus loadItems (see type doc
// comment).
func (c *ChiseledBookshelf) ReadSaveData(tag *nbt.CompoundTag) error {
	raw := int(tag.GetIntOr(chiseledBookshelfTagLastInteractedSlot, 0))
	if raw != 0 {
		slot := blockutils.ChiseledBookshelfSlot(raw - 1)
		c.lastInteractedSlot = &slot
	}
	return nil
}

// WriteSaveData is a port of ChiseledBookshelf::writeSaveData, minus saveItems (see type doc
// comment).
func (c *ChiseledBookshelf) WriteSaveData(tag *nbt.CompoundTag) {
	value := 0
	if c.lastInteractedSlot != nil {
		value = int(*c.lastInteractedSlot) + 1
	}
	tag.SetInt(chiseledBookshelfTagLastInteractedSlot, nbt.IntTag(value))
}
