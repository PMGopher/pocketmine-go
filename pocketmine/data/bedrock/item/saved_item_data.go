package bedrockitem

import (
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
