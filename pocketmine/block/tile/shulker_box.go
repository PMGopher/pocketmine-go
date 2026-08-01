package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const ShulkerBoxTagFacing = "facing"

// ShulkerBox is a port of pocketmine\block\tile\ShulkerBox, minus its inventory/Container half -
// see ContainerComponent's doc comment for why the inventory package can't be imported here.
// Facing, name, and lock are all fully real.
//
// onBlockDestroyedHook is overridden as a no-op in the PHP original (shulker boxes retain their
// contents when destroyed, unlike other containers) - moot here anyway since
// ContainerComponent's onBlockDestroyedHook isn't ported for any container tile (see its doc
// comment), so there's nothing to override.
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
	s.Facing = int(tag.GetByteOr(ShulkerBoxTagFacing, nbt.ByteTag(s.Facing)))
	return nil
}

func (s *ShulkerBox) WriteSaveData(tag *nbt.CompoundTag) {
	s.SaveName(tag)
	tag.SetByte(ShulkerBoxTagFacing, nbt.ByteTag(s.Facing))
}

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
