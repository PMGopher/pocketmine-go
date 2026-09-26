package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const HopperTagTransferCooldown = "TransferCooldown"

// Hopper is a port of pocketmine\block\tile\Hopper. Item transfer isn't implemented in
// PocketMine-MP either (Hopper::onScheduledUpdate is a TODO).
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
	h.loadItems(h, tag)
	h.LoadName(tag)
	h.TransferCooldown = int(tag.GetIntOr(HopperTagTransferCooldown, 0))
	return nil
}

func (h *Hopper) WriteSaveData(tag *nbt.CompoundTag) {
	h.saveItems(h, tag)
	h.SaveName(tag)
	tag.SetInt(HopperTagTransferCooldown, nbt.IntTag(h.TransferCooldown))
}

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (h *Hopper) OnBlockDestroyedHook() { h.dropContents(h) }

// GetInventory is a port of Hopper::getInventory.
func (h *Hopper) GetInventory() Inventory { return h.realInventory(h) }

// GetRealInventory is a port of Hopper::getRealInventory.
func (h *Hopper) GetRealInventory() Inventory { return h.realInventory(h) }

// CloseHook is Hopper::close's removal of the inventory's viewers.
func (h *Hopper) CloseHook() { h.removeAllViewers() }

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (h *Hopper) CopyDataFromItem(item Item) {
	h.TileBase.CopyDataFromItem(item)
	h.ApplyItemCustomName(item)
}
