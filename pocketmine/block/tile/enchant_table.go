package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// EnchantTable is a port of pocketmine\block\tile\EnchantTable.
type EnchantTable struct {
	SpawnableBase
	NameableComponent
}

func NewEnchantTable(world World, pos math.Vector3) *EnchantTable {
	e := &EnchantTable{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}}
	e.Init(e)
	return e
}

func (e *EnchantTable) SaveID() string { return "EnchantTable" }

func (e *EnchantTable) GetDefaultName() string { return "Enchanting Table" }

// ReadSaveData/WriteSaveData are aliased from loadName/saveName in the PHP original (via the
// NameableTrait `as` renaming).
func (e *EnchantTable) ReadSaveData(tag *nbt.CompoundTag) error {
	e.LoadName(tag)
	return nil
}

func (e *EnchantTable) WriteSaveData(tag *nbt.CompoundTag) { e.SaveName(tag) }

func (e *EnchantTable) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	e.NameableComponent.AddAdditionalSpawnData(tag)
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (e *EnchantTable) CopyDataFromItem(item Item) {
	e.TileBase.CopyDataFromItem(item)
	e.ApplyItemCustomName(item)
}

func (e *EnchantTable) GetName() string { return e.NameableComponent.GetName(e) }
