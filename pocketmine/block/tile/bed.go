package tile

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const BedTagColor = "color"

// Bed is a port of pocketmine\block\tile\Bed.
type Bed struct {
	SpawnableBase

	Color blockutils.DyeColor
}

func NewBed(world World, pos math.Vector3) *Bed {
	b := &Bed{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}, Color: blockutils.DyeColorRed}
	b.Init(b)
	return b
}

func (b *Bed) SaveID() string { return "Bed" }

func (b *Bed) GetColor() blockutils.DyeColor { return b.Color }

func (b *Bed) SetColor(color blockutils.DyeColor) { b.Color = color }

// ReadSaveData is a port of Bed::readSaveData. DyeColorIdMap's plain (non-inverted) toId/fromId
// is just the DyeColor's own declaration-order value (0-15, confirmed against the PHP
// registration table in dye_color.go's doc comment), so this reads the byte directly rather than
// porting a whole IntSaveIdMapTrait-based registry for it.
func (b *Bed) ReadSaveData(tag *nbt.CompoundTag) error {
	b.Color = blockutils.DyeColorRed
	if colorTag, ok := tag.GetTag(BedTagColor); ok {
		if byteTag, ok := colorTag.(nbt.ByteTag); ok {
			if int(byteTag) >= int(blockutils.DyeColorWhite) && int(byteTag) <= int(blockutils.DyeColorBlack) {
				b.Color = blockutils.DyeColor(byteTag)
			}
		}
	}
	return nil
}

func (b *Bed) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetByte(BedTagColor, nbt.ByteTag(b.Color))
}

func (b *Bed) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetByte(BedTagColor, nbt.ByteTag(b.Color))
}
