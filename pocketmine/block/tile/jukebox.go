package tile

import (
	"fmt"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Jukebox is a port of pocketmine\block\tile\Jukebox. The record is this package's minimal
// Item, nil for none.
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

const jukeboxTagRecord = "RecordItem" // Item CompoundTag

// RecordStopSoundFunc is `$world->addSound($pos, new RecordStopSound())` (tile's World interface
// has no AddSound). Set by the block package.
var RecordStopSoundFunc func(t Tile)

// isRecord is `$item instanceof Record`.
func isRecord(it Item) bool {
	_, ok := it.(interface{ GetRecordType() blockutils.RecordType })
	return ok
}

// ReadSaveData is a port of Jukebox::readSaveData.
func (j *Jukebox) ReadSaveData(tag *nbt.CompoundTag) error {
	recordTag, ok, err := tag.GetCompoundTag(jukeboxTagRecord)
	if err != nil {
		return err
	}
	if ok {
		record := loadItem(recordTag, fmt.Sprintf("Jukebox (%v) record", j.GetPosition().Vector3))
		if isRecord(record) {
			j.record = record
		}
	}
	return nil
}

// WriteSaveData is a port of Jukebox::writeSaveData.
func (j *Jukebox) WriteSaveData(tag *nbt.CompoundTag) {
	if j.record != nil {
		if recordTag := saveItem(j.record, -1); recordTag != nil {
			tag.SetTag(jukeboxTagRecord, recordTag)
		}
	}
}

// AddAdditionalSpawnData is a port of Jukebox::addAdditionalSpawnData: this is needed for the note
// particles to show on the client side.
func (j *Jukebox) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	if j.record != nil {
		if recordTag := networkItemNbt(j.record); recordTag != nil {
			tag.SetTag(jukeboxTagRecord, recordTag)
		}
	}
}

// OnBlockDestroyedHook is a port of Jukebox::onBlockDestroyedHook.
func (j *Jukebox) OnBlockDestroyedHook() {
	if RecordStopSoundFunc != nil {
		RecordStopSoundFunc(j)
	}
}
