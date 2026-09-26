package bedrockitem

import (
	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
)

const (
	SavedItemDataTagName   = "Name"
	SavedItemDataTagDamage = "Damage"
	SavedItemDataTagBlock  = "Block"
	SavedItemDataTagTag    = "tag"
)

// SavedItemData is a port of pocketmine\data\bedrock\item\SavedItemData: an item's type as saved
// (ID, meta, block state for block items, and extra NBT).
type SavedItemData struct {
	Name  string
	Meta  int
	Block *bedrock.BlockStateData
	Tag   *nbt.CompoundTag
}

func (d SavedItemData) GetName() string                   { return d.Name }
func (d SavedItemData) GetMeta() int                      { return d.Meta }
func (d SavedItemData) GetBlock() *bedrock.BlockStateData { return d.Block }
func (d SavedItemData) GetTag() *nbt.CompoundTag          { return d.Tag }

// ToNbt is a port of SavedItemData::toNbt.
func (d SavedItemData) ToNbt() *nbt.CompoundTag {
	result := nbt.NewCompoundTag()
	result.SetString(SavedItemDataTagName, nbt.StringTag(d.Name))
	result.SetShort(SavedItemDataTagDamage, nbt.ShortTag(d.Meta))
	if d.Block != nil {
		result.SetTag(SavedItemDataTagBlock, d.Block.ToNbt())
	}
	if d.Tag != nil {
		result.SetTag(SavedItemDataTagTag, d.Tag)
	}
	result.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))
	return result
}
