package bedrockitem

import "pocketmine-go/pocketmine/nbt"

// SavedItemStackData tag names, a port of SavedItemStackData::TAG_*.
const (
	SavedItemStackDataTagCount       = "Count"
	SavedItemStackDataTagSlot        = "Slot"
	SavedItemStackDataTagWasPickedUp = "WasPickedUp"
	SavedItemStackDataTagCanPlaceOn  = "CanPlaceOn"
	SavedItemStackDataTagCanDestroy  = "CanDestroy"
)

// SavedItemStackData is a port of pocketmine\data\bedrock\item\SavedItemStackData: an itemstack as
// saved (type data, count and the optional stack-only fields). Slot and WasPickedUp are nil when
// absent.
type SavedItemStackData struct {
	TypeData    SavedItemData
	Count       int
	Slot        *int
	WasPickedUp *bool
	CanPlaceOn  []string
	CanDestroy  []string
}

func (d SavedItemStackData) GetTypeData() SavedItemData { return d.TypeData }
func (d SavedItemStackData) GetCount() int              { return d.Count }
func (d SavedItemStackData) GetSlot() (int, bool) {
	if d.Slot == nil {
		return 0, false
	}
	return *d.Slot, true
}

// ToNbt is a port of SavedItemStackData::toNbt.
func (d SavedItemStackData) ToNbt() *nbt.CompoundTag {
	result := nbt.NewCompoundTag()
	result.SetByte(SavedItemStackDataTagCount, nbt.ByteTag(int8(d.Count)))
	if d.Slot != nil {
		result.SetByte(SavedItemStackDataTagSlot, nbt.ByteTag(int8(*d.Slot)))
	}
	if d.WasPickedUp != nil {
		v := nbt.ByteTag(0)
		if *d.WasPickedUp {
			v = 1
		}
		result.SetByte(SavedItemStackDataTagWasPickedUp, v)
	}
	if len(d.CanPlaceOn) != 0 {
		result.SetTag(SavedItemStackDataTagCanPlaceOn, stringList(d.CanPlaceOn))
	}
	if len(d.CanDestroy) != 0 {
		result.SetTag(SavedItemStackDataTagCanDestroy, stringList(d.CanDestroy))
	}
	return result.Merge(d.TypeData.ToNbt())
}

func stringList(values []string) *nbt.ListTag {
	tags := make([]nbt.Tag, len(values))
	for i, v := range values {
		tags[i] = nbt.StringTag(v)
	}
	list, err := nbt.NewListTag(tags, nbt.TagString)
	if err != nil {
		panic(err)
	}
	return list
}
