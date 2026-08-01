package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// EnderChest is a port of pocketmine\block\tile\EnderChest. Unlike Chest/Barrel/Furnace, this has
// no Container/inventory dependency at all in the PHP original either - an ender chest's contents
// are the opening player's own ender inventory, not something this tile stores - so this is a
// complete, gap-free port.
type EnderChest struct {
	SpawnableBase

	ViewerCount int
}

func NewEnderChest(world World, pos math.Vector3) *EnderChest {
	e := &EnderChest{}
	e.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	e.Init(e)
	return e
}

func (e *EnderChest) SaveID() string { return "EnderChest" }

func (e *EnderChest) GetViewerCount() int { return e.ViewerCount }

// SetViewerCount panics if viewerCount is negative, mirroring the PHP original's
// InvalidArgumentException (a programmer error at the call site).
func (e *EnderChest) SetViewerCount(viewerCount int) {
	if viewerCount < 0 {
		panic("Viewer count cannot be negative")
	}
	e.ViewerCount = viewerCount
}

func (e *EnderChest) ReadSaveData(tag *nbt.CompoundTag) error { return nil }

func (e *EnderChest) WriteSaveData(tag *nbt.CompoundTag) {}

func (e *EnderChest) AddAdditionalSpawnData(tag *nbt.CompoundTag) {}
