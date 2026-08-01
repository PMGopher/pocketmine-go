package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Barrel is a port of pocketmine\block\tile\Barrel, minus its inventory/Container half - see
// ContainerComponent's doc comment for why the inventory package can't be imported here. Name and
// lock are fully real; there's no other state to port.
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
	return nil
}

func (b *Barrel) WriteSaveData(tag *nbt.CompoundTag) { b.SaveName(tag) }

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (b *Barrel) CopyDataFromItem(item Item) {
	b.TileBase.CopyDataFromItem(item)
	b.ApplyItemCustomName(item)
}
