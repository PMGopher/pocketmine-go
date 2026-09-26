package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const ShulkerBoxTagFacing = "facing"

// ShulkerBox is a port of pocketmine\block\tile\ShulkerBox. Unlike other containers it keeps its
// contents when destroyed (its onBlockDestroyedHook is a no-op): they go into the dropped item.
type ShulkerBox struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	Facing int
}

func NewShulkerBox(world World, pos math.Vector3) *ShulkerBox {
	s := &ShulkerBox{Facing: int(math.North)}
	s.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	s.Init(s)
	return s
}

func (s *ShulkerBox) SaveID() string { return "ShulkerBox" }

func (s *ShulkerBox) GetDefaultName() string { return "Shulker Box" }

func (s *ShulkerBox) GetName() string { return s.NameableComponent.GetName(s) }

func (s *ShulkerBox) GetFacing() int { return s.Facing }

func (s *ShulkerBox) SetFacing(facing int) { s.Facing = facing }

func (s *ShulkerBox) ReadSaveData(tag *nbt.CompoundTag) error {
	s.LoadName(tag)
	s.loadItems(s, tag)
	s.Facing = int(tag.GetByteOr(ShulkerBoxTagFacing, nbt.ByteTag(s.Facing)))
	return nil
}

func (s *ShulkerBox) WriteSaveData(tag *nbt.CompoundTag) {
	s.SaveName(tag)
	s.saveItems(s, tag)
	tag.SetByte(ShulkerBoxTagFacing, nbt.ByteTag(s.Facing))
}

// GetCleanedNBT is a port of ShulkerBox::getCleanedNBT: the facing isn't kept in the item.
func (s *ShulkerBox) GetCleanedNBT() *nbt.CompoundTag {
	tag := s.TileBase.GetCleanedNBT()
	if tag != nil {
		tag.RemoveTag(ShulkerBoxTagFacing)
	}
	return tag
}

// GetInventory is a port of ShulkerBox::getInventory.
func (s *ShulkerBox) GetInventory() Inventory { return s.realInventory(s) }

// GetRealInventory is a port of ShulkerBox::getRealInventory.
func (s *ShulkerBox) GetRealInventory() Inventory { return s.realInventory(s) }

// CloseHook is ShulkerBox::close's removal of the inventory's viewers.
func (s *ShulkerBox) CloseHook() { s.removeAllViewers() }

func (s *ShulkerBox) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetByte(ShulkerBoxTagFacing, nbt.ByteTag(s.Facing))
	s.NameableComponent.AddAdditionalSpawnData(tag)
}

// CopyDataFromItem is a port of ShulkerBox::copyDataFromItem, which overrides the usual
// TileBase/NameableComponent behavior entirely: it re-reads the item's whole NBT as save data
// (not just the block-entity-tag subset TileBase.CopyDataFromItem uses), then applies the custom
// name on top.
func (s *ShulkerBox) CopyDataFromItem(item Item) {
	_ = s.ReadSaveData(item.GetNamedTag())
	if item.HasCustomName() {
		s.SetName(item.GetCustomName())
	}
}
