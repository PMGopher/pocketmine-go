package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const HopperTagTransferCooldown = "TransferCooldown"

// Hopper is a port of pocketmine\block\tile\Hopper, minus its inventory/Container half - see
// ContainerComponent's doc comment for why the inventory package can't be imported here.
// TransferCooldown, name, and lock are all fully real; the actual item-transfer simulation would
// need the inventory slots regardless, so it's not ported either way.
type Hopper struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	TransferCooldown int
}

func NewHopper(world World, pos math.Vector3) *Hopper {
	h := &Hopper{}
	h.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	h.Init(h)
	return h
}

func (h *Hopper) SaveID() string { return "Hopper" }

func (h *Hopper) GetDefaultName() string { return "Hopper" }

func (h *Hopper) GetName() string { return h.NameableComponent.GetName(h) }

func (h *Hopper) ReadSaveData(tag *nbt.CompoundTag) error {
	h.LoadName(tag)
	h.TransferCooldown = int(tag.GetIntOr(HopperTagTransferCooldown, 0))
	return nil
}

func (h *Hopper) WriteSaveData(tag *nbt.CompoundTag) {
	h.SaveName(tag)
	tag.SetInt(HopperTagTransferCooldown, nbt.IntTag(h.TransferCooldown))
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (h *Hopper) CopyDataFromItem(item Item) {
	h.TileBase.CopyDataFromItem(item)
	h.ApplyItemCustomName(item)
}
