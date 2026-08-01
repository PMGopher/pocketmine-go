package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Jukebox is a port of pocketmine\block\tile\Jukebox, minus the record's NBT round-trip
// (Item::safeNbtDeserialize/nbtSerialize aren't ported - see Jukebox's [block package] doc
// comment for the same gap) and the network spawn-data translation (TypeConverter isn't ported
// either). The record is instead held as this package's own minimal Item interface, with nil
// standing in for "no record" - same shape as ItemFrame's framed item.
type Jukebox struct {
	SpawnableBase

	record Item
}

func NewJukebox(world World, pos math.Vector3) *Jukebox {
	j := &Jukebox{}
	j.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	j.Init(j)
	return j
}

func (j *Jukebox) SaveID() string { return "Jukebox" }

func (j *Jukebox) GetRecord() (Item, bool) { return j.record, j.record != nil }

func (j *Jukebox) SetRecord(record Item) { j.record = record }

func (j *Jukebox) ReadSaveData(tag *nbt.CompoundTag) error { return nil }

func (j *Jukebox) WriteSaveData(tag *nbt.CompoundTag) {}

// AddAdditionalSpawnData is a no-op - see type doc comment (no item-to-network-NBT translation
// ported, and no real record to read from anyway would be a moot point without it).
func (j *Jukebox) AddAdditionalSpawnData(tag *nbt.CompoundTag) {}
