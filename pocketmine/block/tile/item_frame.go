package tile

import (
	"fmt"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	itemFrameTagItemRotation   = "ItemRotation"
	itemFrameTagItemDropChance = "ItemDropChance"
	itemFrameTagItem           = "Item"
)

// ItemFrame is a port of pocketmine\block\tile\ItemFrame. The framed item is this package's
// minimal Item, nil standing in for VanillaItems::AIR().
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

// SetItem is a port of pocketmine\block\tile\ItemFrame::setItem (the block passes a copy).
func (i *ItemFrame) SetItem(item Item) {
	if isNullItem(item) {
		i.item = nil
	} else {
		i.item = item
	}
}

func (i *ItemFrame) GetItemRotation() int { return i.itemRotation }

func (i *ItemFrame) SetItemRotation(rotation int) { i.itemRotation = rotation }

func (i *ItemFrame) GetItemDropChance() float64 { return i.itemDropChance }

func (i *ItemFrame) SetItemDropChance(chance float64) { i.itemDropChance = chance }

// ReadSaveData is a port of ItemFrame::readSaveData.
func (i *ItemFrame) ReadSaveData(tag *nbt.CompoundTag) error {
	itemTag, ok, err := tag.GetCompoundTag(itemFrameTagItem)
	if err != nil {
		return err
	}
	if ok {
		i.item = loadItem(itemTag, fmt.Sprintf("ItemFrame (%v) framed item", i.GetPosition().Vector3))
		if isNullItem(i.item) {
			i.item = nil
		}
	}
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

// WriteSaveData is a port of ItemFrame::writeSaveData.
func (i *ItemFrame) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetFloat(itemFrameTagItemDropChance, nbt.FloatTag(i.itemDropChance))
	tag.SetFloat(itemFrameTagItemRotation, nbt.FloatTag(i.itemRotation*45))
	if i.item != nil {
		if itemTag := saveItem(i.item, -1); itemTag != nil {
			tag.SetTag(itemFrameTagItem, itemTag)
		}
	}
}

// AddAdditionalSpawnData is a port of ItemFrame::addAdditionalSpawnData.
func (i *ItemFrame) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetFloat(itemFrameTagItemDropChance, nbt.FloatTag(i.itemDropChance))
	tag.SetFloat(itemFrameTagItemRotation, nbt.FloatTag(i.itemRotation*45))
	if i.item != nil {
		if itemTag := networkItemNbt(i.item); itemTag != nil {
			tag.SetTag(itemFrameTagItem, itemTag)
		}
	}
}
