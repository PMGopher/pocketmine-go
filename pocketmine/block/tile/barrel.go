package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Barrel is a port of pocketmine\block\tile\Barrel.
type Barrel struct {
	SpawnableBase
	NameableComponent
	ContainerComponent
}

func NewBarrel(world World, pos math.Vector3) *Barrel {
	b := &Barrel{}
	b.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	b.Init(b)
	return b
}

func (b *Barrel) SaveID() string { return "Barrel" }

func (b *Barrel) GetDefaultName() string { return "Barrel" }

func (b *Barrel) GetName() string { return b.NameableComponent.GetName(b) }

func (b *Barrel) ReadSaveData(tag *nbt.CompoundTag) error {
	b.LoadName(tag)
	b.loadItems(b, tag)
	return nil
}

func (b *Barrel) WriteSaveData(tag *nbt.CompoundTag) {
	b.SaveName(tag)
	b.saveItems(b, tag)
}

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (b *Barrel) OnBlockDestroyedHook() { b.dropContents(b) }

// GetInventory is a port of Barrel::getInventory.
func (b *Barrel) GetInventory() Inventory { return b.realInventory(b) }

// GetRealInventory is a port of Barrel::getRealInventory.
func (b *Barrel) GetRealInventory() Inventory { return b.realInventory(b) }

// CloseHook is Barrel::close's removal of the inventory's viewers.
func (b *Barrel) CloseHook() { b.removeAllViewers() }

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (b *Barrel) CopyDataFromItem(item Item) {
	b.TileBase.CopyDataFromItem(item)
	b.ApplyItemCustomName(item)
}
