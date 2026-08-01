package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	itemFrameTagItemRotation   = "ItemRotation"
	itemFrameTagItemDropChance = "ItemDropChance"
)

// ItemFrame is a port of pocketmine\block\tile\ItemFrame, minus the framed item's NBT round-trip
// (Item::safeNbtDeserialize/nbtSerialize aren't ported - see Jukebox's doc comment for the same
// gap) and the network spawn-data translation (TypeConverter isn't ported either). The framed
// item is instead held as this package's own minimal Item interface, with nil standing in for
// "no item" - there's no Air sentinel available here the way PHP's VanillaItems::AIR() is, since
// tile can't import the item package (see ContainerComponent's doc comment for the same import-
// cycle constraint).
type ItemFrame struct {
	SpawnableBase

	item           Item
	itemRotation   int
	itemDropChance float64
}

func NewItemFrame(world World, pos math.Vector3) *ItemFrame {
	i := &ItemFrame{itemDropChance: 1.0}
	i.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	i.Init(i)
	return i
}

func (i *ItemFrame) SaveID() string { return "ItemFrame" }

func (i *ItemFrame) HasItem() bool { return i.item != nil }

// GetItem is a port of pocketmine\block\tile\ItemFrame::getItem, returning (nil, false) instead
// of PHP's VanillaItems::AIR() when empty - see type doc comment for why.
func (i *ItemFrame) GetItem() (Item, bool) { return i.item, i.item != nil }

// SetItem is a port of pocketmine\block\tile\ItemFrame::setItem. Unlike the PHP original, the
// null-item check ($item->isNull()) is the caller's responsibility (block.ItemFrame's Item has
// IsNull(), tile's minimal Item interface doesn't) - pass nil directly to clear.
func (i *ItemFrame) SetItem(item Item) { i.item = item }

func (i *ItemFrame) GetItemRotation() int { return i.itemRotation }

func (i *ItemFrame) SetItemRotation(rotation int) { i.itemRotation = rotation }

func (i *ItemFrame) GetItemDropChance() float64 { return i.itemDropChance }

func (i *ItemFrame) SetItemDropChance(chance float64) { i.itemDropChance = chance }

func (i *ItemFrame) ReadSaveData(tag *nbt.CompoundTag) error {
	if t, ok := tag.GetTag(itemFrameTagItemRotation); ok {
		if floatTag, ok := t.(nbt.FloatTag); ok {
			i.itemRotation = int(float64(floatTag) / 45)
		} else if byteTag, ok := t.(nbt.ByteTag); ok {
			i.itemRotation = int(byteTag)
		}
	}
	i.itemDropChance = float64(tag.GetFloatOr(itemFrameTagItemDropChance, nbt.FloatTag(i.itemDropChance)))
	return nil
}

func (i *ItemFrame) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetFloat(itemFrameTagItemDropChance, nbt.FloatTag(i.itemDropChance))
	tag.SetFloat(itemFrameTagItemRotation, nbt.FloatTag(i.itemRotation*45))
}

func (i *ItemFrame) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetFloat(itemFrameTagItemDropChance, nbt.FloatTag(i.itemDropChance))
	tag.SetFloat(itemFrameTagItemRotation, nbt.FloatTag(i.itemRotation*45))
}
