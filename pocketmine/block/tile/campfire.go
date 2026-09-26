package tile

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Campfire NBT keys, a port of Campfire's TAG_* constants.
var (
	campfireItemTags        = [4]string{"Item1", "Item2", "Item3", "Item4"}                 // TAG_Compound
	campfireCookingTimeTags = [4]string{"ItemTime1", "ItemTime2", "ItemTime3", "ItemTime4"} // TAG_Int
)

// Campfire item hooks, set by block/inventory (see Inventory).
var (
	// CampfireSlotItemFunc is `$this->inventory->getItem($slot)`, serialized for saving
	// (Item::nbtSerialize), nil for an empty slot.
	CampfireSlotItemFunc func(inv Inventory, slot int) *nbt.CompoundTag
	// CampfireSlotNetworkItemFunc is TypeConverter's getItemTranslator()->toNetworkNbt() for a
	// slot's item, nil for an empty slot.
	CampfireSlotNetworkItemFunc func(inv Inventory, slot int) *nbt.CompoundTag
)

// Campfire is a port of pocketmine\block\tile\Campfire: the items cooking on a campfire and their
// cooking times.
type Campfire struct {
	SpawnableBase
	ContainerComponent

	cookingTimes map[int]int
}

func NewCampfire(world World, pos math.Vector3) *Campfire {
	c := &Campfire{cookingTimes: map[int]int{}}
	c.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	c.Init(c)
	return c
}

func (c *Campfire) SaveID() string { return "Campfire" }

// GetInventory is a port of Campfire::getInventory (a CampfireInventory).
func (c *Campfire) GetInventory() Inventory { return c.realInventory(c) }

// GetRealInventory is a port of Campfire::getRealInventory.
func (c *Campfire) GetRealInventory() Inventory { return c.realInventory(c) }

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (c *Campfire) OnBlockDestroyedHook() { c.dropContents(c) }

// GetCookingTimes is a port of Campfire::getCookingTimes.
func (c *Campfire) GetCookingTimes() map[int]int { return c.cookingTimes }

// SetCookingTimes is a port of Campfire::setCookingTimes.
func (c *Campfire) SetCookingTimes(cookingTimes map[int]int) { c.cookingTimes = cookingTimes }

// ReadSaveData is a port of Campfire::readSaveData.
func (c *Campfire) ReadSaveData(tag *nbt.CompoundTag) error {
	var items []*nbt.CompoundTag
	for slot := 0; slot < 4; slot++ {
		if t, ok := tag.GetTag(campfireItemTags[slot]); ok {
			if itemTag, ok := t.(*nbt.CompoundTag); ok {
				withSlot := itemTag.Clone()
				withSlot.SetByte("Slot", nbt.ByteTag(slot))
				items = append(items, withSlot)
			}
		}
		if t, ok := tag.GetTag(campfireCookingTimeTags[slot]); ok {
			if v, ok := t.(nbt.IntTag); ok {
				c.cookingTimes[slot] = int(v)
			}
		}
	}
	if inv := c.realInventory(c); inv != nil && LoadInventoryItemsFunc != nil {
		LoadInventoryItemsFunc(inv, items, fmt.Sprintf("Campfire (%v)", c.position.Vector3))
	}
	return nil
}

// WriteSaveData is a port of Campfire::writeSaveData.
func (c *Campfire) WriteSaveData(tag *nbt.CompoundTag) {
	inv := c.realInventory(c)
	if inv == nil || CampfireSlotItemFunc == nil {
		return
	}
	for slot := 0; slot < 4; slot++ {
		if itemTag := CampfireSlotItemFunc(inv, slot); itemTag != nil {
			tag.SetTag(campfireItemTags[slot], itemTag)
			if t, ok := c.cookingTimes[slot]; ok {
				tag.SetInt(campfireCookingTimeTags[slot], nbt.IntTag(t))
			}
		}
	}
}

// AddAdditionalSpawnData is a port of Campfire::addAdditionalSpawnData.
func (c *Campfire) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	inv := c.realInventory(c)
	if inv == nil || CampfireSlotNetworkItemFunc == nil {
		return
	}
	for slot := 0; slot < 4; slot++ {
		if itemTag := CampfireSlotNetworkItemFunc(inv, slot); itemTag != nil {
			tag.SetTag(campfireItemTags[slot], itemTag)
		}
	}
}
